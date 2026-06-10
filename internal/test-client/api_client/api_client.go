package api_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"simple-server/internal/model"
	"simple-server/internal/util"

	"github.com/google/uuid"
)

type ApiClient struct {
	baseUrl string
}

func NewApiClient(baseUrl string) *ApiClient {
	return &ApiClient{
		baseUrl: baseUrl,
	}

}

// запрос списка заметок
func (c *ApiClient) FetchAllNotes() ([]model.Note, error) {
	fmt.Println("GET /notes")
	resp, err := http.Get(fmt.Sprintf("%s/notes", c.baseUrl))
	if err != nil {
		return nil, err
	}
	defer util.CloseResponseBody(resp)
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(resp.Status)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("body: %s", string(raw))
	}

	var notesList []model.Note
	if err := util.DecodeJson(raw, &notesList); err != nil {
		return nil, err
	}
	return notesList, nil
}

func (c *ApiClient) FetchNotesByHeader(header string) ([]model.Note, error) {
	params := url.Values{}
	params.Add("header", header)
	path := "/notes?" + params.Encode()
	fmt.Println("GET " + path)

	resp, err := http.Get(c.baseUrl + path)
	if err != nil {
		return nil, err
	}
	defer util.CloseResponseBody(resp)
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(resp.Status)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("body: %s", string(raw))
	}

	var notes []model.Note
	if err := util.DecodeJson(raw, &notes); err != nil {
		return nil, err
	}
	return notes, nil
}

// запрос заметки
func (c *ApiClient) FetchNoteByID(ID uuid.UUID) (*model.Note, error) {
	fmt.Println("GET /note/" + ID.String())

	resp, err := http.Get(fmt.Sprintf("%s/note/%s", c.baseUrl, ID.String()))
	if err != nil {
		return nil, err
	}
	defer util.CloseResponseBody(resp)
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(resp.Status)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("body: %s", string(raw))
	}

	var note model.Note
	if err := util.DecodeJson(raw, &note); err != nil {
		return nil, err
	}
	return &note, nil
}

// запрос добавления новой заметки
func (c *ApiClient) AddNote(note *model.Note) (*model.Note, error) {
	jsonBody, _ := json.Marshal(note)

	fmt.Printf("POST /note \n%s", jsonBody)
	resp, err := http.Post(fmt.Sprintf("%s/note", c.baseUrl), "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer util.CloseResponseBody(resp)
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(resp.Status)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("body: %s", string(raw))
	}

	var addedNote model.Note
	if err := util.DecodeJson(raw, &addedNote); err != nil {
		return nil, fmt.Errorf("json unmarshaling error: %w", err)
	}

	return &addedNote, nil
}

// запрос изменения заметки
func (c *ApiClient) UpdateNote(note *model.Note) error {
	updateNote := model.UpdateNoteRequestBody{
		Header: note.Header,
		Body:   note.Body,
	}
	jsonBody, _ := json.Marshal(updateNote)

	path := fmt.Sprintf("/note/%s", url.PathEscape(note.ID.String()))
	fmt.Printf("PUT %s \n%s\n", path, jsonBody)
	req, _ := http.NewRequest(http.MethodPut, c.baseUrl+path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer util.CloseResponseBody(resp)
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
func (c *ApiClient) DeleteNote(ID uuid.UUID) error {
	fmt.Printf("DELETE /note/%s", ID.String())
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/note/%s", c.baseUrl, ID.String()), nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	fmt.Println(resp.Status)

	defer util.CloseResponseBody(resp)
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
func (c *ApiClient) Convert(amount float64, baseCurrency string, targetCurrencies []string) (uuid.UUID, error) {
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
		return uuid.Nil, err
	}
	defer util.CloseResponseBody(resp)
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return uuid.Nil, err
	}
	fmt.Println(resp.Status)
	if resp.StatusCode != 200 {
		return uuid.Nil, fmt.Errorf("body: %s", string(raw))
	}

	var result model.ConvertCurrencyResponse
	if err := util.DecodeJson(raw, &result); err != nil {
		return uuid.Nil, err
	}
	return result.TaskID, nil
}

func (c *ApiClient) FetchTaskStatus(taskID uuid.UUID) (*model.TaskStatus, error) {
	fmt.Printf("GET /task/%s/status\n", taskID.String())
	resp, err := http.Get(fmt.Sprintf("%s/task/%s/status", c.baseUrl, taskID.String()))
	if err != nil {
		return nil, err
	}
	defer util.CloseResponseBody(resp)
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(resp.Status)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("body: %s", string(raw))
	}
	var status model.TaskStatus
	if err := util.DecodeJson(raw, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

func (c *ApiClient) FetchTask(taskID uuid.UUID) (*model.Task, error) {
	fmt.Printf("GET /task/%s\n", taskID.String())
	resp, err := http.Get(fmt.Sprintf("%s/task/%s", c.baseUrl, taskID.String()))
	if err != nil {
		return nil, err
	}
	defer util.CloseResponseBody(resp)
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(resp.Status)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("body: %s", string(raw))
	}
	var task model.Task
	if err := util.DecodeJson(raw, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (c *ApiClient) DeleteTask(taskID uuid.UUID) error {
	fmt.Printf("DELETE /task/%s\n", taskID.String())
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/task/%s", c.baseUrl, taskID.String()), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer util.CloseResponseBody(resp)
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
