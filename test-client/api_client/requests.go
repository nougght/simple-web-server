package api_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type Note struct {
	Header string `json:"header"`
	Body   string `json:"body"`
}

type ApiClient struct {
	baseUrl string
}

func NewApiClient(baseUrl string) *ApiClient {
	return &ApiClient{
		baseUrl: baseUrl,
	}

}

// запрос списка заметок
func (c *ApiClient) FetchAllNotes() ([]Note, error) {
	fmt.Println("GET /notes")
	resp, err := http.Get(fmt.Sprintf("%s/notes", c.baseUrl))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(resp.Status)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("body: %s", string(raw))
	}

	var notesList []Note
	if err := json.Unmarshal(raw, &notesList); err != nil {
		return nil, err
	}
	return notesList, nil
}

// запрос заметки
func (c *ApiClient) FetchNote(header string) (*Note, error) {
	fmt.Println("GET /notes/" + header)

	resp, err := http.Get(fmt.Sprintf("%s/notes/%s", c.baseUrl, header))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(resp.Status)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("body: %s", string(raw))
	}

	var note Note
	if err := json.Unmarshal(raw, &note); err != nil {
		return nil, err
	}
	return &note, nil
}

// запрос добавления новой заметки
func (c *ApiClient) AddNote(note *Note) error {
	jsonBody, _ := json.Marshal(note)

	fmt.Printf("POST /notes \n%s", jsonBody)
	resp, err := http.Post(fmt.Sprintf("%s/notes", c.baseUrl), "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(resp.Status)
	if resp.StatusCode != 200 {
		return fmt.Errorf("body: %s", string(raw))
	}

	return nil
}

// запрос изменения заметки
func (c *ApiClient) UpdateNote(note *Note) error {
	jsonBody, _ := json.Marshal(*note)

	path := fmt.Sprintf("/notes/%s", url.PathEscape(note.Header))
	fmt.Printf("PUT %s \n%s\n", path, jsonBody)
	req, _ := http.NewRequest(http.MethodPut, c.baseUrl+path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(resp.Status)
	if resp.StatusCode != 200 {
		return fmt.Errorf("body: %s", string(raw))
	}

	return nil
}

// запрос удаления заметки
func (c *ApiClient) DeleteNote(header string) error {
	fmt.Printf("DELETE /notes/%s \n%s", header)
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/notes", c.baseUrl), nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	fmt.Println(resp.Status)

	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("body: %s", string(raw))
	}

	return nil
}

// запрос конвертации валют
func (c *ApiClient) Convert(amount float64, baseCurrency string, targetCurrencies []string) (map[string]float64, error) {
	params := url.Values{}
	if amount != 0 {
		params.Add("amount", strconv.FormatFloat(amount, 'f', 3, 64))
	}
	if baseCurrency != "" {
		params.Add("base", baseCurrency)
	}
	if len(targetCurrencies) != 0 {
		for _, val := range targetCurrencies {
			params.Add("currencies", val)
		}
	}
	path := "/currency?" + params.Encode()
	fmt.Printf("GET %s\n", path)
	resp, err := http.Get(c.baseUrl + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(resp.Status)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("body: %s", string(raw))
	}

	var result map[string]float64
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil

}
