# About

Lominus is a tool written in Go to automatically sync Luminus files onto your local storage for easy access at a fixed interval of your choice. It is designed to run on Windows, macOS and Linux operating systems.

# Features

Lominus removes the hassle to download (or redownload) whenever files are uploaded (or updated), resulting in necessary manual cleanup and organisation of the files. And many more!

- Keeps your Luminus files organised
- Automatic download of all files from Luminus Files for all modules
- Automatic update of files when files are reuploaded on Luminus
- System notification sync status
- Dark/Light mode (based on your system's theme)
- System tray icon (Windows only)

To be implemented:

- Notification when new grades are released in gradebooks
- User-defined webhook for Telegram/Notion/custom integrations

# Getting Started

## Installation

Standalone programs for Windows, macOS and Linux are available. Download the program from [releases](https://github.com/beebeeoii/lominus/releases).

## Building

You may choose to clone and compile the program for your operating system manually.

### Prerequisites

1. [Go](https://go.dev/dl/)

2. gcc

### Build

1. Ensure Go is set in you system env var.

2. Navigate to the directory where you cloned and execute `go install`.

3. Install [fyne](https://developer.fyne.io/index.html) via `go get fyne.io/fyne/v2/cmd/fyne`

4. Finally, build and compile. Depending on your system os:

    - For windows: `fyne package -os windows`.

    - For macOS: `fyne package -os darwin`.

    - For Linux: `fyne package -os linux`.

## API

Lominus can also be used as an API. Please visit documentations (coming soon) for more details.

# Screenshots

Login Info             |  Preferences
:-------------------------:|:-------------------------:
![image](./screenshots/ubuntu-1.png)  |  ![image](./screenshots/ubuntu-2.png)
![image](./screenshots/mac-1.png)  |  ![image](./screenshots/mac-2.png)
![image](./screenshots/win-1.png)  |  ![image](./screenshots/win-2.png)

# Contributing

Lominus is far from perfect. All contributions, regardless large or small, are highly encouraged as this would help to make Lominus better. Please submit an [issue](https://github.com/beebeeoii/lominus/issues) or fork this repo and submit a [pull request](https://github.com/beebeeoii/lominus/pulls) if necessary.

# FAQ

1. I am unable to verify my Luminus login credentials.

    - Please ensure that you have trusted the Certificate used by Luminus, **especially on macOS**.
