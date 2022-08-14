// Package api provides functions that link up and communicate with LMS servers,
// such as Canvas and Luminus (probably removed in near future).
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// MultimediaChannel struct is the datapack for containing details about every multimedia channel in a module.
type MultimediaChannel struct {
	Id          string
	Name        string
	MediaCount  int
	LastUpdated int64
}

// MultimediaVideo struct is the datapack for containing details about every multimedia video in a module.
type MultimediaVideo struct {
	Id         string
	Name       string
	FolderId   string // Also known as Channel Id
	FolderName string // Also known as Channel Name
	M3u8Url    string
}

type panaptoCredentials struct {
	aspAuth   string
	csrfToken string
	folderId  string
}

type PanaptoVideoRawResponse struct {
	D PanaptoVideoResponse `json:"d"`
}

type PanaptoVideoResponse struct {
	Results []map[string]interface{} `json:"Results"`
	Total   int                      `json:"TotalNumber"`
}

const MULTIMEMDIA_CHANNEL_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/multimedia/?populate=contentSummary&ParentID=%s"
const LTI_DATA_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/lti/Launch/mediaweb?context_id=%s&returnURL=https://luminus.nus.edu.sg/iframe/lti-return/mediaweb"

const PANAPTO_AUTH_URL_ENDPOINT = "https://mediaweb.ap.panopto.com/Panopto/LTI/LTI.aspx"
const PANAPTO_VIDEOS_URL_ENDPOINT = "https://mediaweb.ap.panopto.com/Panopto/Services/Data.svc/GetSessions"
const PANAPTO_VIDEO_DELIVERY_URL_ENDPOINT = "https://mediaweb.ap.panopto.com/Panopto/Pages/Viewer/DeliveryInfo.aspx"

const PANAPTO_ASPAUTH_KEY = ".ASPXAUTH"
const PANAPTO_CSRF_KEY = "csrfToken"

// getMultimediaChannelFieldsRequired is a helper function that returns a constant array with fields that a multimedia channel element
// returned by Luminus needs.
func getMultimediaChannelFieldsRequired() []string {
	return []string{"access", "id", "name", "mediaCount", "lastUpdatedDate"}
}

func (req MultimediaChannelRequest) GetMultimediaChannels() ([]MultimediaChannel, error) {
	var multimediaChannels []MultimediaChannel

	rawResponse := RawResponse{}
	err := req.Request.GetRawResponse(&rawResponse)
	if err != nil {
		return multimediaChannels, err
	}

	for _, content := range rawResponse.Data {
		if !IsResponseValid(getMultimediaChannelFieldsRequired(), content) {
			continue
		}

		if _, exists := content["access"]; exists {
			lastUpdated := int64(-1)
			lastUpdatedTime, err := time.Parse(time.RFC3339, content["lastUpdatedDate"].(string))
			if err != nil {
				return multimediaChannels, err
			}
			lastUpdated = lastUpdatedTime.Unix()

			multimediaChannel := MultimediaChannel{
				Id:          content["id"].(string),
				Name:        content["name"].(string),
				MediaCount:  int(content["mediaCount"].(float64)),
				LastUpdated: lastUpdated,
			}

			multimediaChannels = append(multimediaChannels, multimediaChannel)
		}
	}

	return multimediaChannels, nil
}

func (req MultimediaVideoRequest) GetMultimediaVideos() ([]MultimediaVideo, error) {
	panaptoAuthReqBody := url.Values{}

	var multimediaVideos []MultimediaVideo

	ltiDataResponse := LTIDataResponse{}
	err := req.Request.GetRawResponse(&ltiDataResponse)
	if err != nil {
		return multimediaVideos, err
	}

	for _, content := range ltiDataResponse.DataItems {
		panaptoAuthReqBody.Set(content["key"].(string), content["value"].(string))
	}

	panaptoCredentials, getPanaptoCredentialsErr := getPanaptoCredentials(panaptoAuthReqBody)
	if getPanaptoCredentialsErr != nil {
		return multimediaVideos, err
	}

	return getPanaptoVideos(panaptoCredentials)
}

func getPanaptoCredentials(data url.Values) (panaptoCredentials, error) {
	var panaptoCredentials panaptoCredentials
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	panaptoAuthReq, panaptoAuthReqErr := http.NewRequest(POST, PANAPTO_AUTH_URL_ENDPOINT, strings.NewReader(data.Encode()))
	if panaptoAuthReqErr != nil {
		return panaptoCredentials, panaptoAuthReqErr
	}

	panaptoAuthReq.Header.Add("Content-Type", CONTENT_TYPE_FORM)
	panaptoAuthReq.Header.Add("User-Agent", USER_AGENT)

	authRes, authResErr := client.Do(panaptoAuthReq)
	if authResErr != nil {
		return panaptoCredentials, authResErr
	}

	if len(authRes.Cookies()) < 2 {
		return panaptoCredentials, fmt.Errorf("unable to get Panapto credentials - invalid auth res cookies received")
	}

	for _, cookie := range authRes.Cookies() {
		if cookie.Name == PANAPTO_ASPAUTH_KEY {
			panaptoCredentials.aspAuth = cookie.Value
		}

		if cookie.Name == PANAPTO_CSRF_KEY {
			panaptoCredentials.csrfToken = cookie.Value
		}
	}

	location := authRes.Header.Get("Location")

	indexStart := strings.Index(location, "folderID%3D") + 11
	indexEnd := strings.Index(location, "%26isLTIEmbed")
	panaptoCredentials.folderId = location[indexStart:indexEnd]

	return panaptoCredentials, nil
}

func getPanaptoVideos(credentials panaptoCredentials) ([]MultimediaVideo, error) {
	var multimediaVideos []MultimediaVideo

	jsonBody := map[string]map[string]string{
		"queryParameters": {
			"folderID": credentials.folderId,
		},
	}
	jsonValue, _ := json.Marshal(jsonBody)
	client := &http.Client{}

	panaptoVideoReq, panaptoVideoReqErr := http.NewRequest(POST, PANAPTO_VIDEOS_URL_ENDPOINT, bytes.NewBuffer(jsonValue))
	if panaptoVideoReqErr != nil {
		return multimediaVideos, panaptoVideoReqErr
	}

	panaptoVideoReq.Header.Add("Content-Type", CONTENT_TYPE_JSON)
	panaptoVideoReq.Header.Add("User-Agent", USER_AGENT)
	panaptoVideoReq.AddCookie(&http.Cookie{
		Name:  ".ASPXAUTH",
		Value: credentials.aspAuth,
	})
	panaptoVideoReq.AddCookie(&http.Cookie{
		Name:  "csrfToken",
		Value: credentials.csrfToken,
	})

	videoRes, videoResErr := client.Do(panaptoVideoReq)
	if videoResErr != nil {
		return multimediaVideos, videoResErr
	}

	body, bodyErr := ioutil.ReadAll(videoRes.Body)
	if bodyErr != nil {
		return multimediaVideos, bodyErr
	}

	var rawResponse PanaptoVideoRawResponse
	json.Unmarshal(body, &rawResponse)

	for _, video := range rawResponse.D.Results {
		multimediaVideo := MultimediaVideo{
			Id:         video["SessionID"].(string),
			Name:       video["SessionName"].(string),
			FolderId:   video["FolderID"].(string),
			FolderName: video["FolderName"].(string),
			M3u8Url:    video["IosVideoUrl"].(string),
		}

		multimediaVideos = append(multimediaVideos, multimediaVideo)
	}

	return multimediaVideos, nil
}

// func (video MultimediaVideo) Download(dir string) error {
// 	m3u8.
// }
