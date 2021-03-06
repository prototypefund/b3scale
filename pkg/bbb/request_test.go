package bbb

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
)

func TestParamsString(t *testing.T) {
	var p Params
	if p.String() != "" {
		t.Error("expected empty string")
	}

	tests := map[string]Params{
		// Test parameter ordering
		"a=23&b=true&c=foo": Params{
			"c":        "foo",
			"a":        "23",
			"b":        "true",
			"checksum": "fff0000000000fff",
		},

		// URL-safe encoding
		"name=Meeting+Name": Params{
			"name": "Meeting Name",
		},
	}

	for expected, params := range tests {
		result := params.String()
		if result != expected {
			t.Error("Unexpected result:", result)
		}
	}
}

func TestParamsGetMeetingID(t *testing.T) {
	p1 := Params{
		"meetingID": "someMeetingID",
		"foo":       "bar",
	}
	p2 := Params{
		"foo": "bar",
	}

	// Found
	id, ok := p1.MeetingID()
	if !ok {
		t.Error("expected meetingID")
	}
	if id != "someMeetingID" {
		t.Error("Unexpected meetingID:", id)
	}

	// Not Found
	id, ok = p2.MeetingID()
	if ok {
		t.Error("did not expect meetingID:", id)
	}
}

func TestSign(t *testing.T) {
	// We use the example from the api documentation.
	// However as we encode our parameters with a deterministic
	// order, different to the example, we will end up with
	// a different checksum.
	backend := &Backend{
		Secret: "639259d4-9dd8-4b25-bf01-95f9567eaf4b",
	}
	req := &Request{
		Backend:  backend,
		Resource: "create",
		Params: Params{
			"name":        "Test Meeting",
			"meetingID":   "abc123",
			"attendeePW":  "111222",
			"moderatorPW": "333444",
		},
	}

	checksum := req.Sign()
	if checksum != "94ec9a89c7dc53af01537aef9f8ecbae5e95cd7f37cd4bf18101b976a4a8b097" {
		t.Error("Unexpected checksum:", checksum)
	}
}

func TestVerify(t *testing.T) {
	// We use the example from the api documentation, now
	// for validating against a frontend secret.
	// order, different to the example, we will end up with
	// a different checksum.
	frontend := &Frontend{
		Secret: "639259d4-9dd8-4b25-bf01-95f9567eaf4b",
	}
	params := Params{
		"name":        "Test Meeting",
		"meetingID":   "abc123",
		"attendeePW":  "111222",
		"moderatorPW": "333444",
	}
	req := &Request{
		Frontend: frontend,
		Resource: "create",
		Request: &http.Request{
			URL: &url.URL{
				RawQuery: params.String() + "&checksum=r3m0v3M3",
			},
		},
		Params:   params,
		Checksum: "0b89c2ebcfefb76772cbcf19386c33561f66f6ae",
	}

	// Success
	err := req.Verify()
	if err != nil {
		t.Error(err)
	}

	// Error
	req.Checksum = "foob4r"
	err = req.Verify()
	if err == nil {
		t.Error("Expected a checksum error.")
	}
}

func TestString(t *testing.T) {
	// Request create to backend
	backend := &Backend{
		Host:   "https://bbbackend",
		Secret: "639259d4-9dd8-4b25-bf01-95f9567eaf4b",
	}
	req := &Request{
		Backend:  backend,
		Resource: "create",
		Params: Params{
			"name":        "Test Meeting",
			"meetingID":   "abc123",
			"attendeePW":  "111222",
			"moderatorPW": "333444",
		},
	}

	// Call stringer
	reqURL := req.URL()
	expected := "https://bbbackend/create" +
		"?attendeePW=111222&meetingID=abc123" +
		"&moderatorPW=333444&name=Test+Meeting&" +
		"checksum=94ec9a89c7dc53af01537aef9f8ecbae5e95cd7f37cd4bf18101b976a4a8b097"
	if reqURL != expected {
		t.Error("Unexpected request URL:", reqURL)
	}

	// No params
	req.Params = Params{}
	reqURL = req.URL()
	expected = "https://bbbackend/create" +
		"?checksum=272c9555258496a3f19c5ad8f599af2a4ebec031381ff1e37b34842c42c12284"
	if reqURL != expected {
		t.Error("Unexpected request URL:", reqURL)
	}
}

func TestUnAndMarshalURLSafe(t *testing.T) {
	req := JoinRequest(Params{
		"meetingID": "abcd1235789-foo",
		"userID":    "optional",
		"checksum":  "12342",
	})
	reqURL, _ := url.Parse("/bbb/frontend/join?foo")
	req.Request.URL = reqURL

	enc := req.MarshalURLSafe()
	req1, err := UnmarshalURLSafeRequest(enc)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(req1)
	id, _ := req1.Params.MeetingID()
	if id != "abcd1235789-foo" {

	}
}

func TestDecodeURLSafeRequest(t *testing.T) {
	req := JoinRequest(Params{
		"meetingID": "abcd1235789-foo",
		"userID":    "optional",
		"checksum":  "12342",
	})

	reqURL, _ := url.Parse("/bbb/frontend/join?foo=42")
	hdr := http.Header{}
	hdr.Set("content-type", "application/test")
	req.Request.Header = hdr
	req.Request.URL = reqURL

	data := req.MarshalURLSafe()
	payload := make([]byte, base64.RawURLEncoding.DecodedLen(len(data)))
	if _, err := base64.RawURLEncoding.Decode(payload, data); err != nil {
		t.Fatal(err)
	}

	var r interface{}
	if err := json.Unmarshal(payload, &r); err != nil {
		t.Fatal(err)
	}

	req1 := decodeURLSafeRequest(r)
	if req1 == nil {
		t.Error("decode failed.")
	}
	/*
		if req1.Request.Header.Get("content-type") != "application/test" {
			t.Error("unexpected http header", req1.Request.Header)
		}
	*/
	if req1.Request.URL.Path != "/bbb/frontend/join" {
		t.Error("unexpected path", req1.Request.URL.Path)
	}
	if req1.Request.URL.Query().Get("foo") != "42" {
		t.Error("unexpected query")
	}
}
