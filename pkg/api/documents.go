package api

import "fmt"

type Folder struct {
	Id           string
	Name         string
	Subfolder    bool
	downloadable bool
}

type Document struct {
	Id   string
	Name string
}

const FOLDER_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/files/?populate=totalFileCount,subFolderCount,TotalSize&ParentID=%s"
const DOC_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/files/%s/file?populate=Creator,lastUpdatedUser,comment"

func (req Request) GetAllFolders() ([]Folder, error) { //consider making this method private
	folder := []Folder{}

	RawResponse, err := req.GetRawResponse()

	if err != nil {
		return folder, err
	}

	for _, content := range RawResponse.Data {

		if _, ok := content["access"]; ok { // only folder that can be accessed will be placed in folders slice

			newFolder := Folder{
				content["id"].(string),
				content["name"].(string),
				content["subFolderCount"].(float64) > 0,
				content["isActive"].(bool) && !content["allowUpload"].(bool), // downloadable = active folder + does not allow uploads
			}
			folder = append(folder, newFolder)
		}
	}
	return folder, nil
}

func (req Request) GetAllFileNames() ([]Document, error) {
	documents := []Document{}
	fols, err := req.GetAllFolders()
	if err != nil {
		return documents, err
	}

	for _, f := range fols {
		newReq := Request{
			Url:       fmt.Sprintf(FOLDER_URL_ENDPOINT, f.Id),
			JwtToken:  req.JwtToken,
			UserAgent: USER_AGENT,
		}
		docs, err := newReq.getRootFiles(f)

		if err != nil {
			return documents, err
		}

		documents = append(documents, docs...)
	}
	return documents, nil
}

func (req Request) getRootFiles(folder Folder) ([]Document, error) {
	documents := []Document{}

	if !folder.Subfolder && folder.downloadable {
		newReq := Request{
			Url:       fmt.Sprintf(DOC_URL_ENDPOINT, folder.Id),
			JwtToken:  req.JwtToken,
			UserAgent: USER_AGENT,
		}
		RawResponse, err := newReq.GetRawResponse()

		if err != nil {
			return documents, err
		}
		for _, content := range RawResponse.Data {

			newDoc := Document{
				content["id"].(string),
				content["name"].(string),
			}
			documents = append(documents, newDoc)
		}
	} else {
		newReq := Request{
			Url:       fmt.Sprintf(FOLDER_URL_ENDPOINT, folder.Id),
			JwtToken:  req.JwtToken,
			UserAgent: USER_AGENT,
		}
		rawFols, err := newReq.GetAllFolders()

		if err != nil {
			return documents, err
		}

		for _, fol := range rawFols {
			newreq := Request{
				Url:       fmt.Sprintf(FOLDER_URL_ENDPOINT, fol.Id),
				JwtToken:  req.JwtToken,
				UserAgent: USER_AGENT,
			}
			docs, err := newreq.getRootFiles(fol)

			if err != nil {
				return documents, err
			}

			documents = append(documents, docs...)
		}
	}
	return documents, nil
}
