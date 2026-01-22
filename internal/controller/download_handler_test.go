package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"youclips/internal/entities"
	"youclips/internal/service"
)

// In-memory repo for tests
type memRepo struct {
	seq   int
	items map[int]*entities.Clip
}

func newMemRepo() *memRepo { return &memRepo{items: make(map[int]*entities.Clip)} }

func (m *memRepo) Create(ctx context.Context, clip *entities.Clip) error {
	m.seq++
	clip.ID = m.seq
	m.items[clip.ID] = clone(clip)
	return nil
}
func (m *memRepo) GetByID(ctx context.Context, id int) (*entities.Clip, error) {
	c := m.items[id]
	if c == nil { return nil, nil }
	return clone(c), nil
}
func (m *memRepo) List(ctx context.Context, limit, offset int) ([]*entities.Clip, error) {
	var res []*entities.Clip
	for i := 1; i <= m.seq; i++ {
		if c := m.items[i]; c != nil { res = append(res, clone(c)) }
	}
	// simulate pagination
	end := offset + limit
	if offset > len(res) { return []*entities.Clip{}, nil }
	if end > len(res) { end = len(res) }
	return res[offset:end], nil
}
func (m *memRepo) Count(ctx context.Context) (int, error) { return len(m.items), nil }
func (m *memRepo) UpdateStatus(ctx context.Context, id int, status entities.ClipStatus) error {
	if c := m.items[id]; c != nil { c.Status = status }
	return nil
}
func (m *memRepo) Update(ctx context.Context, clip *entities.Clip) error {
	if c := m.items[clip.ID]; c != nil { m.items[clip.ID] = clone(clip) }
	return nil
}

func clone(c *entities.Clip) *entities.Clip {
	cc := *c
	return &cc
}

// Test processor that marks clip completed, writes file, and persists in repo
type testProcessor struct{ dir string; repo *memRepo }

func (p *testProcessor) ProcessClip(ctx context.Context, clip *entities.Clip) error {
	name := filepath.Join(p.dir, "clip_"+strconv.Itoa(clip.ID))
	if clip.Format == entities.FormatVideo { name += ".mp4" } else { name += ".mp3" }
	if err := os.WriteFile(name, []byte("data"), 0644); err != nil { return err }
	clip.Title = "test"
	clip.FilePath = name
	clip.Size = int64(len("data"))
	clip.Status = entities.StatusCompleted
	// persist updated clip in mem repo
	_ = p.repo.Update(ctx, clip)
	return nil
}

func setupHandler(t *testing.T) *ClipHandler {
	t.Helper()
	repo := newMemRepo()
	dir := t.TempDir()
	proc := &testProcessor{dir: dir, repo: repo}
	svc := service.NewClipService(repo, proc)
	return NewClipHandler(svc)
}

func TestPOSTClips_Create(t *testing.T) {
	h := setupHandler(t)
	reqBody := `{"url":"http://x","start_time":0,"end_time":5,"format":"video"}`
	req := httptest.NewRequest(http.MethodPost, "/clips", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Clips(rec, req)
	if rec.Code != http.StatusCreated { t.Fatalf("expected 201, got %d", rec.Code) }
	var resp CreateClipResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil { t.Fatal(err) }
	if resp.ID == 0 || resp.Status == "" { t.Fatalf("invalid response: %+v", resp) }
}

func TestGETClips_List(t *testing.T) {
	h := setupHandler(t)
	// create 3 clips
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/clips", strings.NewReader(`{"url":"u","start_time":0,"end_time":1,"format":"video"}`))
		rec := httptest.NewRecorder()
		h.Clips(rec, req)
	}
	listReq := httptest.NewRequest(http.MethodGet, "/clips?page=1&limit=2", nil)
	listRec := httptest.NewRecorder()
	h.Clips(listRec, listReq)
	if listRec.Code != http.StatusOK { t.Fatalf("expected 200, got %d", listRec.Code) }
	var resp ListClipsResponse
	json.Unmarshal(listRec.Body.Bytes(), &resp)
	if resp.Total != 3 || len(resp.Clips) != 2 || resp.Page != 1 { t.Fatalf("unexpected list resp: %+v", resp) }
}

func TestGETClipByID_Get(t *testing.T) {
	h := setupHandler(t)
	// create
	reqBody := `{"url":"http://x","start_time":0,"end_time":5,"format":"audio"}`
	req := httptest.NewRequest(http.MethodPost, "/clips", strings.NewReader(reqBody))
	rec := httptest.NewRecorder()
	h.Clips(rec, req)
	var created CreateClipResponse
	json.Unmarshal(rec.Body.Bytes(), &created)
	// get
	getReq := httptest.NewRequest(http.MethodGet, "/clips/"+strconv.Itoa(created.ID), nil)
	getRec := httptest.NewRecorder()
	h.ClipByID(getRec, getReq)
	if getRec.Code != http.StatusOK { t.Fatalf("expected 200, got %d", getRec.Code) }
	var resp ClipResponse
	json.Unmarshal(getRec.Body.Bytes(), &resp)
	if resp.ID != created.ID { t.Fatalf("expected id %d, got %d", created.ID, resp.ID) }
	if resp.Status != string(entities.StatusProcessing) { t.Fatalf("expected completed, got %s", resp.Status) }
}
func TestGETClipByID_NotFound(t *testing.T) {
	h := setupHandler(t)
	
	//create 
	reqBody := `{"url":"http://x","start_time":0,"end_time":5,"format":"audio"}`
	req := httptest.NewRequest(http.MethodPost, "/clips", strings.NewReader(reqBody))
	rec := httptest.NewRecorder()
	h.Clips(rec, req)
	// get non-existent
	getReq := httptest.NewRequest(http.MethodGet, "/clips/999", nil)
	getRec := httptest.NewRecorder()
	h.ClipByID(getRec, getReq)
	if getRec.Code != http.StatusNotFound { t.Fatalf("expected 404, got %d", getRec.Code) }

	}
func TestGETClipByID_Download(t *testing.T) {
	h := setupHandler(t)
	// create clip
	req := httptest.NewRequest(http.MethodPost, "/clips", strings.NewReader(`{"url":"u","start_time":0,"end_time":1,"format":"video"}`))
	rec := httptest.NewRecorder()
	h.Clips(rec, req)
	var created CreateClipResponse
	json.Unmarshal(rec.Body.Bytes(), &created)
	
	// directly update status in mem repo to simulate completion
	// (not ideal in real app, but fine for unit test)
	// find clip and mark completed
	// since we don't have direct access to repo here, rely on processor having persisted completion
	// download
		dlReq := httptest.NewRequest(http.MethodGet, "/clips/"+strconv.Itoa(created.ID)+"/download", nil)
	dlRec := httptest.NewRecorder()
	h.ClipByID(dlRec, dlReq)
	if dlRec.Code != http.StatusOK { t.Fatalf("expected 200, got %d", dlRec.Code) }
	ct := dlRec.Header().Get("Content-Type")
	if ct != "video/mp4" { t.Fatalf("unexpected content-type: %s", ct) }
}
