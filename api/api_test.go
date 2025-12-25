package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/parkan/go-hauk/config"
	"github.com/parkan/go-hauk/store"
)

func testServer() (*Server, *store.Memory) {
	mem := store.NewMemory()
	cfg := &config.Config{
		PublicURL:      "https://example.com/",
		MaxDuration:    86400,
		MinInterval:    1,
		MaxCachedPts:   3,
		MaxShownPts:    100,
		LinkStyle:      0,
		AllowLinkReq:   true,
		PasswordHash:   "$2a$10$LerNFYkUU3ZZrNHhamISZeDK8afdExOwDKbyTaUECDOLa1rV4iN.O", // "test"
		AuthMethod:     config.AuthPassword,
		RateLimitAuth:  10000,
		RateLimitAdopt: 10000,
		TrustProxy:     true,
	}
	return NewServer(cfg, mem), mem
}

func postForm(srv http.Handler, path string, data url.Values) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", path, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	return w
}

func getPath(srv http.Handler, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", path, nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	return w
}

func TestCreateSession(t *testing.T) {
	srv, _ := testServer()

	t.Run("missing fields", func(t *testing.T) {
		w := postForm(srv, "/api/create.php", url.Values{})
		if !strings.Contains(w.Body.String(), "Missing data!") {
			t.Errorf("expected missing data error, got: %s", w.Body.String())
		}
	})

	t.Run("bad password", func(t *testing.T) {
		w := postForm(srv, "/api/create.php", url.Values{
			"dur": {"3600"},
			"int": {"5"},
			"pwd": {"wrong"},
		})
		if !strings.Contains(w.Body.String(), "Incorrect password!") {
			t.Errorf("expected password error, got: %s", w.Body.String())
		}
	})

	t.Run("invalid duration", func(t *testing.T) {
		w := postForm(srv, "/api/create.php", url.Values{
			"dur": {"abc"},
			"int": {"5"},
			"pwd": {"test"},
		})
		if !strings.Contains(w.Body.String(), "Invalid duration!") {
			t.Errorf("expected duration error, got: %s", w.Body.String())
		}
	})

	t.Run("duration too long", func(t *testing.T) {
		w := postForm(srv, "/api/create.php", url.Values{
			"dur": {"999999"},
			"int": {"5"},
			"pwd": {"test"},
		})
		if !strings.Contains(w.Body.String(), "exceeds maximum") {
			t.Errorf("expected max duration error, got: %s", w.Body.String())
		}
	})

	t.Run("interval too short", func(t *testing.T) {
		w := postForm(srv, "/api/create.php", url.Values{
			"dur": {"3600"},
			"int": {"0.1"},
			"pwd": {"test"},
		})
		if !strings.Contains(w.Body.String(), "too short") {
			t.Errorf("expected interval error, got: %s", w.Body.String())
		}
	})

	t.Run("solo share", func(t *testing.T) {
		w := postForm(srv, "/api/create.php", url.Values{
			"dur": {"3600"},
			"int": {"5"},
			"pwd": {"test"},
			"mod": {"0"},
		})

		lines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
		if len(lines) < 4 || lines[0] != "OK" {
			t.Fatalf("expected OK response, got: %s", w.Body.String())
		}
		if lines[1] == "" {
			t.Error("expected session ID")
		}
		if !strings.HasPrefix(lines[2], "https://example.com/") {
			t.Errorf("expected view link, got: %s", lines[2])
		}
	})

	t.Run("group share", func(t *testing.T) {
		w := postForm(srv, "/api/create.php", url.Values{
			"dur": {"3600"},
			"int": {"5"},
			"pwd": {"test"},
			"mod": {"1"},
			"nic": {"alice"},
		})

		lines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
		if len(lines) < 5 || lines[0] != "OK" {
			t.Fatalf("expected OK response with pin, got: %s", w.Body.String())
		}
		// group share returns: OK, sid, url, pin, id
		if lines[3] == "" {
			t.Error("expected PIN for group share")
		}
	})

	t.Run("e2e encrypted", func(t *testing.T) {
		w := postForm(srv, "/api/create.php", url.Values{
			"dur":  {"3600"},
			"int":  {"5"},
			"pwd":  {"test"},
			"mod":  {"0"},
			"e2e":  {"1"},
			"salt": {"abcd1234"},
		})

		lines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
		if lines[0] != "OK" {
			t.Errorf("expected OK, got: %s", w.Body.String())
		}
	})

	t.Run("e2e without salt", func(t *testing.T) {
		w := postForm(srv, "/api/create.php", url.Values{
			"dur": {"3600"},
			"int": {"5"},
			"pwd": {"test"},
			"mod": {"0"},
			"e2e": {"1"},
		})
		if !strings.Contains(w.Body.String(), "Missing data!") {
			t.Errorf("expected error for e2e without salt, got: %s", w.Body.String())
		}
	})

	t.Run("e2e with group disallowed", func(t *testing.T) {
		w := postForm(srv, "/api/create.php", url.Values{
			"dur":  {"3600"},
			"int":  {"5"},
			"pwd":  {"test"},
			"mod":  {"1"},
			"nic":  {"alice"},
			"e2e":  {"1"},
			"salt": {"abcd"},
		})
		if !strings.Contains(w.Body.String(), "not supported for group") {
			t.Errorf("expected e2e group error, got: %s", w.Body.String())
		}
	})
}

func TestPostLocation(t *testing.T) {
	srv, mem := testServer()

	// create a session first
	w := postForm(srv, "/api/create.php", url.Values{
		"dur": {"3600"},
		"int": {"5"},
		"pwd": {"test"},
		"mod": {"0"},
	})
	lines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
	if lines[0] != "OK" {
		t.Fatalf("setup failed: %s", w.Body.String())
	}
	sid := lines[1]
	shareID := lines[3]

	t.Run("missing fields", func(t *testing.T) {
		w := postForm(srv, "/api/post.php", url.Values{})
		if !strings.Contains(w.Body.String(), "Missing data!") {
			t.Errorf("expected missing data, got: %s", w.Body.String())
		}
	})

	t.Run("invalid session", func(t *testing.T) {
		w := postForm(srv, "/api/post.php", url.Values{
			"sid":  {"invalid"},
			"lat":  {"51.5"},
			"lon":  {"-0.1"},
			"time": {"1234567890"},
		})
		if !strings.Contains(w.Body.String(), "Session expired!") {
			t.Errorf("expected session expired, got: %s", w.Body.String())
		}
	})

	t.Run("invalid coordinates", func(t *testing.T) {
		w := postForm(srv, "/api/post.php", url.Values{
			"sid":  {sid},
			"lat":  {"999"},
			"lon":  {"-0.1"},
			"time": {"1234567890"},
		})
		if !strings.Contains(w.Body.String(), "Invalid location!") {
			t.Errorf("expected invalid location, got: %s", w.Body.String())
		}
	})

	t.Run("valid post", func(t *testing.T) {
		w := postForm(srv, "/api/post.php", url.Values{
			"sid":  {sid},
			"lat":  {"51.5074"},
			"lon":  {"-0.1278"},
			"time": {"1234567890.123"},
			"spd":  {"5.5"},
			"acc":  {"10"},
		})

		lines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
		if lines[0] != "OK" {
			t.Errorf("expected OK, got: %s", w.Body.String())
		}
		if !strings.Contains(lines[2], shareID) {
			t.Errorf("expected share ID in response, got: %s", lines[2])
		}
	})

	t.Run("expired session", func(t *testing.T) {
		// clear and don't set up session
		mem.Clear()

		w := postForm(srv, "/api/post.php", url.Values{
			"sid":  {sid},
			"lat":  {"51.5"},
			"lon":  {"-0.1"},
			"time": {"1234567890"},
		})
		if !strings.Contains(w.Body.String(), "Session expired!") {
			t.Errorf("expected expired, got: %s", w.Body.String())
		}
	})
}

func TestFetch(t *testing.T) {
	srv, _ := testServer()

	// create session and post location
	w := postForm(srv, "/api/create.php", url.Values{
		"dur": {"3600"},
		"int": {"5"},
		"pwd": {"test"},
		"mod": {"0"},
	})
	lines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
	sid := lines[1]
	shareID := lines[3]

	// post a location
	postForm(srv, "/api/post.php", url.Values{
		"sid":  {sid},
		"lat":  {"51.5074"},
		"lon":  {"-0.1278"},
		"time": {"1234567890.123"},
	})

	t.Run("missing id", func(t *testing.T) {
		w := getPath(srv, "/api/fetch.php")
		if !strings.Contains(w.Body.String(), "Invalid session!") {
			t.Errorf("expected invalid session, got: %s", w.Body.String())
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		w := getPath(srv, "/api/fetch.php?id=invalid")
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got: %d", w.Code)
		}
	})

	t.Run("valid fetch", func(t *testing.T) {
		w := getPath(srv, "/api/fetch.php?id="+shareID)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got: %d", w.Code)
		}

		var resp soloResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("json decode error: %v", err)
		}

		if resp.Type != 0 {
			t.Errorf("expected type 0, got: %d", resp.Type)
		}
		if resp.Interval != 5 {
			t.Errorf("expected interval 5, got: %f", resp.Interval)
		}
		if len(resp.Points) != 1 {
			t.Errorf("expected 1 point, got: %d", len(resp.Points))
		}
		if len(resp.Points) > 0 {
			pt := resp.Points[0]
			if len(pt) < 3 {
				t.Fatal("point has too few fields")
			}
			if pt[0].(float64) != 51.5074 {
				t.Errorf("expected lat 51.5074, got: %v", pt[0])
			}
		}
	})

	t.Run("fetch with since filter", func(t *testing.T) {
		// post another location with later timestamp
		postForm(srv, "/api/post.php", url.Values{
			"sid":  {sid},
			"lat":  {"52.0"},
			"lon":  {"0.0"},
			"time": {"1234567900.0"},
		})

		// fetch with since after first point
		w := getPath(srv, "/api/fetch.php?id="+shareID+"&since=1234567895")
		var resp soloResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("json decode error: %v", err)
		}

		if len(resp.Points) != 1 {
			t.Errorf("expected 1 point after since filter, got: %d", len(resp.Points))
		}
	})
}

func TestGroupShare(t *testing.T) {
	srv, _ := testServer()

	// create group share
	w := postForm(srv, "/api/create.php", url.Values{
		"dur": {"3600"},
		"int": {"5"},
		"pwd": {"test"},
		"mod": {"1"},
		"nic": {"alice"},
	})
	lines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
	if lines[0] != "OK" {
		t.Fatalf("group create failed: %s", w.Body.String())
	}
	aliceSid := lines[1]
	pin := lines[3]
	shareID := lines[4]

	// post alice's location
	postForm(srv, "/api/post.php", url.Values{
		"sid":  {aliceSid},
		"lat":  {"51.5"},
		"lon":  {"-0.1"},
		"time": {"1234567890"},
	})

	t.Run("join group", func(t *testing.T) {
		w := postForm(srv, "/api/create.php", url.Values{
			"dur": {"3600"},
			"int": {"5"},
			"pwd": {"test"},
			"mod": {"2"},
			"nic": {"bob"},
			"pin": {pin},
		})
		lines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
		if lines[0] != "OK" {
			t.Fatalf("join failed: %s", w.Body.String())
		}
		bobSid := lines[1]

		// post bob's location
		postForm(srv, "/api/post.php", url.Values{
			"sid":  {bobSid},
			"lat":  {"52.0"},
			"lon":  {"0.0"},
			"time": {"1234567891"},
		})
	})

	t.Run("join invalid pin", func(t *testing.T) {
		w := postForm(srv, "/api/create.php", url.Values{
			"dur": {"3600"},
			"int": {"5"},
			"pwd": {"test"},
			"mod": {"2"},
			"nic": {"eve"},
			"pin": {"999999"},
		})
		if !strings.Contains(w.Body.String(), "Invalid group PIN!") {
			t.Errorf("expected invalid pin error, got: %s", w.Body.String())
		}
	})

	t.Run("fetch group", func(t *testing.T) {
		w := getPath(srv, "/api/fetch.php?id="+shareID)
		var resp groupResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("decode error: %v", err)
		}

		if resp.Type != 1 {
			t.Errorf("expected type 1 (group), got: %d", resp.Type)
		}
		if len(resp.Points) != 2 {
			t.Errorf("expected 2 participants, got: %d", len(resp.Points))
		}
		if _, ok := resp.Points["alice"]; !ok {
			t.Error("expected alice in points")
		}
		if _, ok := resp.Points["bob"]; !ok {
			t.Error("expected bob in points")
		}
	})
}

func TestStopSession(t *testing.T) {
	srv, _ := testServer()

	w := postForm(srv, "/api/create.php", url.Values{
		"dur": {"3600"},
		"int": {"5"},
		"pwd": {"test"},
	})
	lines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
	sid := lines[1]
	shareID := lines[3]

	t.Run("stop valid", func(t *testing.T) {
		w := postForm(srv, "/api/stop.php", url.Values{
			"sid": {sid},
		})
		if !strings.Contains(w.Body.String(), "OK") {
			t.Errorf("expected OK, got: %s", w.Body.String())
		}

		// verify session is gone
		w = getPath(srv, "/api/fetch.php?id="+shareID)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404 after stop, got: %d", w.Code)
		}
	})

	t.Run("stop invalid", func(t *testing.T) {
		w := postForm(srv, "/api/stop.php", url.Values{
			"sid": {"invalid"},
		})
		// should still return OK (idempotent)
		if !strings.Contains(w.Body.String(), "OK") {
			t.Errorf("expected OK even for invalid, got: %s", w.Body.String())
		}
	})
}

func TestDynamic(t *testing.T) {
	srv, _ := testServer()

	w := getPath(srv, "/dynamic.js.php")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got: %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "TILE_URI") {
		t.Error("expected TILE_URI in dynamic.js")
	}
	if !strings.Contains(body, "VELOCITY_UNIT") {
		t.Error("expected VELOCITY_UNIT in dynamic.js")
	}
}

func TestVersionHeader(t *testing.T) {
	srv, _ := testServer()

	w := getPath(srv, "/api/fetch.php?id=test")
	version := w.Header().Get("X-Hauk-Version")
	if version == "" {
		t.Error("expected X-Hauk-Version header")
	}
	if !strings.Contains(version, "-go") {
		t.Errorf("expected version to contain '-go', got: %s", version)
	}
}

func TestNewLink(t *testing.T) {
	srv, _ := testServer()

	// create session
	w := postForm(srv, "/api/create.php", url.Values{
		"dur": {"3600"},
		"int": {"5"},
		"pwd": {"test"},
	})
	lines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
	sid := lines[1]

	t.Run("missing sid", func(t *testing.T) {
		w := postForm(srv, "/api/new-link.php", url.Values{})
		if !strings.Contains(w.Body.String(), "Missing data!") {
			t.Errorf("expected missing data, got: %s", w.Body.String())
		}
	})

	t.Run("invalid session", func(t *testing.T) {
		w := postForm(srv, "/api/new-link.php", url.Values{
			"sid": {"invalid"},
		})
		if !strings.Contains(w.Body.String(), "Session expired!") {
			t.Errorf("expected session expired, got: %s", w.Body.String())
		}
	})

	t.Run("create new link", func(t *testing.T) {
		w := postForm(srv, "/api/new-link.php", url.Values{
			"sid": {sid},
		})
		lines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
		if lines[0] != "OK" {
			t.Errorf("expected OK, got: %s", w.Body.String())
		}
		if len(lines) < 3 {
			t.Fatalf("expected 3 lines, got: %d", len(lines))
		}
		if !strings.HasPrefix(lines[1], "https://example.com/") {
			t.Errorf("expected view link, got: %s", lines[1])
		}
	})

	t.Run("create adoptable link", func(t *testing.T) {
		w := postForm(srv, "/api/new-link.php", url.Values{
			"sid": {sid},
			"ado": {"1"},
		})
		if !strings.Contains(w.Body.String(), "OK") {
			t.Errorf("expected OK, got: %s", w.Body.String())
		}
	})
}

func TestAdopt(t *testing.T) {
	srv, _ := testServer()

	// create group share (owner)
	w := postForm(srv, "/api/create.php", url.Values{
		"dur": {"3600"},
		"int": {"5"},
		"pwd": {"test"},
		"mod": {"1"},
		"nic": {"group-owner"},
	})
	ownerLines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
	ownerSid := ownerLines[1]
	groupPin := ownerLines[3]

	// create adoptable solo share
	w = postForm(srv, "/api/create.php", url.Values{
		"dur": {"3600"},
		"int": {"5"},
		"pwd": {"test"},
		"mod": {"0"},
		"ado": {"1"},
	})
	soloLines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
	soloSid := soloLines[1]
	soloShareID := soloLines[3]

	t.Run("missing fields", func(t *testing.T) {
		w := postForm(srv, "/api/adopt.php", url.Values{})
		if !strings.Contains(w.Body.String(), "Missing data!") {
			t.Errorf("expected missing data, got: %s", w.Body.String())
		}
	})

	t.Run("invalid session", func(t *testing.T) {
		w := postForm(srv, "/api/adopt.php", url.Values{
			"sid": {"invalid"},
			"nic": {"adopter"},
			"aid": {soloShareID},
			"pin": {groupPin},
		})
		if !strings.Contains(w.Body.String(), "Session expired!") {
			t.Errorf("expected session expired, got: %s", w.Body.String())
		}
	})

	t.Run("share not found", func(t *testing.T) {
		w := postForm(srv, "/api/adopt.php", url.Values{
			"sid": {ownerSid},
			"nic": {"adopter"},
			"aid": {"nonexistent"},
			"pin": {groupPin},
		})
		if !strings.Contains(w.Body.String(), "Share not found!") {
			t.Errorf("expected share not found, got: %s", w.Body.String())
		}
	})

	t.Run("successful adopt", func(t *testing.T) {
		w := postForm(srv, "/api/adopt.php", url.Values{
			"sid": {soloSid},
			"nic": {"adopted-user"},
			"aid": {soloShareID},
			"pin": {groupPin},
		})
		if !strings.Contains(w.Body.String(), "OK") {
			t.Errorf("expected OK, got: %s", w.Body.String())
		}
	})

	t.Run("unauthorized adopt", func(t *testing.T) {
		// create another adoptable share
		w = postForm(srv, "/api/create.php", url.Values{
			"dur": {"3600"},
			"int": {"5"},
			"pwd": {"test"},
			"mod": {"0"},
			"ado": {"1"},
		})
		lines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
		anotherShareID := lines[3]

		// try to adopt with wrong session
		w = postForm(srv, "/api/adopt.php", url.Values{
			"sid": {ownerSid},
			"nic": {"attacker"},
			"aid": {anotherShareID},
			"pin": {groupPin},
		})
		if !strings.Contains(w.Body.String(), "Not authorized!") {
			t.Errorf("expected not authorized, got: %s", w.Body.String())
		}
	})

	t.Run("non-adoptable share", func(t *testing.T) {
		// create non-adoptable share
		w = postForm(srv, "/api/create.php", url.Values{
			"dur": {"3600"},
			"int": {"5"},
			"pwd": {"test"},
			"mod": {"0"},
			"ado": {"0"},
		})
		lines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
		nonAdoptableSid := lines[1]
		nonAdoptableID := lines[3]

		w = postForm(srv, "/api/adopt.php", url.Values{
			"sid": {nonAdoptableSid},
			"nic": {"adopter"},
			"aid": {nonAdoptableID},
			"pin": {groupPin},
		})
		if !strings.Contains(w.Body.String(), "not allowed") {
			t.Errorf("expected not allowed error, got: %s", w.Body.String())
		}
	})
}

func TestEncryptedSession(t *testing.T) {
	srv, _ := testServer()

	// create encrypted session
	w := postForm(srv, "/api/create.php", url.Values{
		"dur":  {"3600"},
		"int":  {"5"},
		"pwd":  {"test"},
		"e2e":  {"1"},
		"salt": {"abc123"},
	})
	lines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
	sid := lines[1]
	shareID := lines[3]

	t.Run("post encrypted location", func(t *testing.T) {
		w := postForm(srv, "/api/post.php", url.Values{
			"sid":  {sid},
			"lat":  {"encrypted_lat_data"},
			"lon":  {"encrypted_lon_data"},
			"time": {"encrypted_time_data"},
			"iv":   {"initialization_vector"},
		})
		if !strings.Contains(w.Body.String(), "OK") {
			t.Errorf("expected OK, got: %s", w.Body.String())
		}
	})

	t.Run("fetch encrypted has salt", func(t *testing.T) {
		w := getPath(srv, "/api/fetch.php?id="+shareID)
		var resp soloResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("json decode error: %v", err)
		}

		if !resp.Encrypted {
			t.Error("expected encrypted flag")
		}
		if resp.Salt != "abc123" {
			t.Errorf("expected salt abc123, got: %s", resp.Salt)
		}
	})
}
