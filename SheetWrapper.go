package main

import (
        "C"
        "context"
        "encoding/json"
        "fmt"
        "log"
        "net/http"
        "os"

        "golang.org/x/oauth2"
        "golang.org/x/oauth2/google"
        "google.golang.org/api/option"
        "google.golang.org/api/sheets/v4"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(tokenPath string, config *oauth2.Config) *http.Client {
        // The file token.json stores the user's access and refresh tokens, and is
        // created automatically when the authorization flow completes for the first
        // time.
        tok, err := tokenFromFile(tokenPath)
        if err != nil {
                tok = getTokenFromWeb(config)
                saveToken(tokenPath, tok)
        }
        return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
        authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
        fmt.Printf("Go to the following link in your browser then type the "+
                "authorization code: \n%v\n", authURL)

        var authCode string
        if _, err := fmt.Scan(&authCode); err != nil {
                log.Fatalf("Unable to read authorization code: %v", err)
        }

        tok, err := config.Exchange(context.TODO(), authCode)
        if err != nil {
                log.Fatalf("Unable to retrieve token from web: %v", err)
        }
        return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
        f, err := os.Open(file)
        if err != nil {
                return nil, err
        }
        defer f.Close()
        tok := &oauth2.Token{}
        err = json.NewDecoder(f).Decode(tok)
        return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
        fmt.Printf("Saving credential file to: %s\n", path)
        f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
        if err != nil {
                log.Fatalf("Unable to cache oauth token: %v", err)
        }
        defer f.Close()
        json.NewEncoder(f).Encode(token)
}

//export readTable
func readTable(credPath string, tokenPath string, sheetId string, sheetName string) (C.int, []C.int, []*C.char) {
    ctx := context.Background()
    b, err := os.ReadFile(credPath)
    if err != nil {
        log.Fatalf("Unable to read client secret file: %v", err)
    }

    // If modifying these scopes, delete your previously saved token.json.
    config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
    if err != nil {
            log.Fatalf("Unable to parse client secret file to config: %v", err)
    }
    client := getClient(tokenPath, config)

    srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
    if err != nil {
            log.Fatalf("Unable to retrieve Sheets client: %v", err)
    }

    // Prints the names and majors of students in a sample spreadsheet:
    // https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
    resp, err := srv.Spreadsheets.Values.Get(sheetId, sheetName).MajorDimension("COLUMNS").Do()
    if err != nil {
            log.Fatalf("Unable to retrieve data from sheet: %v", err)
    }

    if len(resp.Values) == 0 {
            return 0, nil, nil
    }

    var l_NumRow C.int = C.int(len(resp.Values))
    var l_NumColumn []C.int
    var l_Cells []*C.char
    for _, row := range resp.Values {
        l_NumColumn = append(l_NumColumn, C.int(len(row)))
		for _, col := range row{
			l_Cells = append(l_Cells, C.CString(fmt.Sprintf("%s", col)))
		}
	}
    return l_NumRow, l_NumColumn, l_Cells
}

func main() {} 