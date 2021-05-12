package bbb

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
)

const (
	// RetSuccess is the success return code
	RetSuccess = "SUCCESS"

	// RetFailed is the failure return code
	RetFailed = "FAILED"
)

// Response interface
type Response interface {
	Marshal() ([]byte, error)

	Header() http.Header
	SetHeader(http.Header)

	Status() int
	SetStatus(int)
}

// A RawResponse contains just the blob received from
// the server.
type RawResponse struct {
	body   []byte
	header http.Header
	status int
}

// SetBody sets the RawResponse body
func (res *RawResponse) SetBody(body []byte) {
	res.body = body
}

// Marshal returns the response data
func (res *RawResponse) Marshal() ([]byte, error) {
	return res.body, nil
}

// Header returns the response http header
func (res *RawResponse) Header() http.Header {
	if res.header == nil {
		res.header = make(http.Header)
		res.header.Add("Content-Type", "text/html")
	}
	return res.header
}

// SetHeader sets the http response header
func (res *RawResponse) SetHeader(h http.Header) {
	res.header = h
}

// Status returns the reponse status. Default is 200 OK.
func (res *RawResponse) Status() int {
	if res.status == 0 {
		return http.StatusOK
	}
	return res.status
}

// SetStatus sets thes response status
func (res *RawResponse) SetStatus(s int) {
	res.status = s
}

// A XMLResponse from the server
type XMLResponse struct {
	XMLName    xml.Name `xml:"response"`
	Returncode string   `xml:"returncode"`
	Message    string   `xml:"message,omitempty"`
	MessageKey string   `xml:"messageKey,omitempty"`
	Version    string   `xml:"version,omitempty"`

	header http.Header
	status int
}

// Marshal a XMLResponse to XML
func (res *XMLResponse) Marshal() ([]byte, error) {
	data, err := xml.Marshal(res)
	return data, err
}

// Header returns the HTTP response headers
func (res *XMLResponse) Header() http.Header {
	if res.header == nil {
		res.header = make(http.Header)
		res.header.Add("Content-Type", "application/xml")
	}
	return res.header
}

// SetHeader sets the HTTP response headers
func (res *XMLResponse) SetHeader(h http.Header) {
	res.header = h
}

// Status returns the HTTP response status code
func (res *XMLResponse) Status() int {
	return res.status
}

// SetStatus sets the HTTP response status code
func (res *XMLResponse) SetStatus(s int) {
	res.status = s
}

// CreateResponse is the resonse for the `create` API resource.
type CreateResponse struct {
	*XMLResponse
	*Meeting
}

// UnmarshalCreateResponse decodes the resonse XML data.
func UnmarshalCreateResponse(data []byte) (*CreateResponse, error) {
	res := &CreateResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal a CreateResponse to XML
func (res *CreateResponse) Marshal() ([]byte, error) {
	data, err := xml.Marshal(res)
	return data, err
}

// Header returns the HTTP response headers
func (res *CreateResponse) Header() http.Header {
	return res.XMLResponse.Header()
}

// SetHeader sets the HTTP response headers
func (res *CreateResponse) SetHeader(h http.Header) {
	res.XMLResponse.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *CreateResponse) Status() int {
	return res.XMLResponse.Status()
}

// SetStatus sets the HTTP response status code
func (res *CreateResponse) SetStatus(s int) {
	res.XMLResponse.SetStatus(s)
}

// JoinResponse of the join resource.
// WARNING: the join response might be a html page without
// any meaningful data.
type JoinResponse struct {
	*XMLResponse
	MeetingID    string `xml:"meeting_id"`
	UserID       string `xml:"user_id"`
	AuthToken    string `xml:"auth_token"`
	SessionToken string `xml:"session_token"`
	URL          string `xml:"url"`
}

// UnmarshalJoinResponse decodes the serialized XML data
func UnmarshalJoinResponse(data []byte) (*JoinResponse, error) {
	res := &JoinResponse{}
	err := xml.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Marshal encodes a JoinResponse as XML
func (res *JoinResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// Header returns the HTTP response headers
func (res *JoinResponse) Header() http.Header {
	return res.XMLResponse.Header()
}

// SetHeader sets the HTTP response headers
func (res *JoinResponse) SetHeader(h http.Header) {
	res.XMLResponse.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *JoinResponse) Status() int {
	return res.XMLResponse.Status()
}

// SetStatus sets the HTTP response status code
func (res *JoinResponse) SetStatus(s int) {
	res.XMLResponse.SetStatus(s)
}

// IsMeetingRunningResponse is a meeting status resonse
type IsMeetingRunningResponse struct {
	*XMLResponse
	Running bool `xml:"running"`
}

// UnmarshalIsMeetingRunningResponse decodes the XML data.
func UnmarshalIsMeetingRunningResponse(
	data []byte,
) (*IsMeetingRunningResponse, error) {
	res := &IsMeetingRunningResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal a IsMeetingRunningResponse to XML
func (res *IsMeetingRunningResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// Header returns the HTTP response headers
func (res *IsMeetingRunningResponse) Header() http.Header {
	return res.XMLResponse.Header()
}

// SetHeader sets the HTTP response headers
func (res *IsMeetingRunningResponse) SetHeader(h http.Header) {
	res.XMLResponse.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *IsMeetingRunningResponse) Status() int {
	return res.XMLResponse.Status()
}

// SetStatus sets the HTTP response status code
func (res *IsMeetingRunningResponse) SetStatus(s int) {
	res.XMLResponse.SetStatus(s)
}

// EndResponse is the resonse of the end resource
type EndResponse struct {
	*XMLResponse
}

// UnmarshalEndResponse decodes the xml resonse
func UnmarshalEndResponse(data []byte) (*EndResponse, error) {
	res := &EndResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal EndResponse to XML
func (res *EndResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// Header returns the HTTP response headers
func (res *EndResponse) Header() http.Header {
	return res.XMLResponse.Header()
}

// SetHeader sets the HTTP response headers
func (res *EndResponse) SetHeader(h http.Header) {
	res.XMLResponse.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *EndResponse) Status() int {
	return res.XMLResponse.Status()
}

// SetStatus sets the HTTP response status code
func (res *EndResponse) SetStatus(s int) {
	res.XMLResponse.SetStatus(s)
}

// GetMeetingInfoResponse contains detailed meeting information
type GetMeetingInfoResponse struct {
	*XMLResponse
	*Meeting
}

// UnmarshalGetMeetingInfoResponse decodes the xml response
func UnmarshalGetMeetingInfoResponse(
	data []byte,
) (*GetMeetingInfoResponse, error) {
	res := &GetMeetingInfoResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal GetMeetingInfoResponse to XML
func (res *GetMeetingInfoResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// Header returns the HTTP response headers
func (res *GetMeetingInfoResponse) Header() http.Header {
	return res.XMLResponse.Header()
}

// SetHeader sets the HTTP response headers
func (res *GetMeetingInfoResponse) SetHeader(h http.Header) {
	res.XMLResponse.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *GetMeetingInfoResponse) Status() int {
	return res.XMLResponse.Status()
}

// SetStatus sets the HTTP response status code
func (res *GetMeetingInfoResponse) SetStatus(s int) {
	res.XMLResponse.SetStatus(s)
}

// GetMeetingsResponse contains a list of meetings.
type GetMeetingsResponse struct {
	*XMLResponse
	Meetings []*Meeting `xml:"meetings>meeting"`
}

// UnmarshalGetMeetingsResponse decodes the xml response
func UnmarshalGetMeetingsResponse(
	data []byte,
) (*GetMeetingsResponse, error) {
	res := &GetMeetingsResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal serializes the response as XML
func (res *GetMeetingsResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// Header returns the HTTP response headers
func (res *GetMeetingsResponse) Header() http.Header {
	return res.XMLResponse.Header()
}

// SetHeader sets the HTTP response headers
func (res *GetMeetingsResponse) SetHeader(h http.Header) {
	res.XMLResponse.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *GetMeetingsResponse) Status() int {
	return res.XMLResponse.Status()
}

// SetStatus sets the HTTP response status code
func (res *GetMeetingsResponse) SetStatus(s int) {
	res.XMLResponse.SetStatus(s)
}

// GetRecordingsResponse is the response of the getRecordings resource
type GetRecordingsResponse struct {
	*XMLResponse
	Recordings []*Recording `xml:"recordings>recording"`
}

// UnmarshalGetRecordingsResponse deserializes the response XML
func UnmarshalGetRecordingsResponse(
	data []byte,
) (*GetRecordingsResponse, error) {
	res := &GetRecordingsResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal a GetRecordingsResponse to XML
func (res *GetRecordingsResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// Header returns the HTTP response headers
func (res *GetRecordingsResponse) Header() http.Header {
	return res.XMLResponse.Header()
}

// SetHeader sets the HTTP response headers
func (res *GetRecordingsResponse) SetHeader(h http.Header) {
	res.XMLResponse.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *GetRecordingsResponse) Status() int {
	return res.XMLResponse.Status()
}

// SetStatus sets the HTTP response status code
func (res *GetRecordingsResponse) SetStatus(s int) {
	res.XMLResponse.SetStatus(s)
}

// PublishRecordingsResponse indicates if the recordings
// were published. This also has the potential for
// tasks failed successfully.
// Also the endpoint is designed badly because you can send
// a set of recordings and receive just a single published
// true or false.
type PublishRecordingsResponse struct {
	*XMLResponse
	Published bool `xml:"published"`
}

// UnmarshalPublishRecordingsResponse decodes the XML response
func UnmarshalPublishRecordingsResponse(
	data []byte,
) (*PublishRecordingsResponse, error) {
	res := &PublishRecordingsResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal a publishRecodingsResponse to XML
func (res *PublishRecordingsResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// Header returns the HTTP response headers
func (res *PublishRecordingsResponse) Header() http.Header {
	return res.XMLResponse.Header()
}

// SetHeader sets the HTTP response headers
func (res *PublishRecordingsResponse) SetHeader(h http.Header) {
	res.XMLResponse.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *PublishRecordingsResponse) Status() int {
	return res.XMLResponse.Status()
}

// SetStatus sets the HTTP response status code
func (res *PublishRecordingsResponse) SetStatus(s int) {
	res.XMLResponse.SetStatus(s)
}

// DeleteRecordingsResponse indicates if the recording
// was correctly deleted. Might fail successfully.
// Same crap as with the publish resource
type DeleteRecordingsResponse struct {
	*XMLResponse
	Deleted bool `xml:"deleted"`
}

// UnmarshalDeleteRecordingsResponse decodes XML resource response
func UnmarshalDeleteRecordingsResponse(
	data []byte,
) (*DeleteRecordingsResponse, error) {
	res := &DeleteRecordingsResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal encodes the delete recordings response as XML
func (res *DeleteRecordingsResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// Header returns the HTTP response headers
func (res *DeleteRecordingsResponse) Header() http.Header {
	return res.XMLResponse.Header()
}

// SetHeader sets the HTTP response headers
func (res *DeleteRecordingsResponse) SetHeader(h http.Header) {
	res.XMLResponse.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *DeleteRecordingsResponse) Status() int {
	return res.XMLResponse.Status()
}

// SetStatus sets the HTTP response status code
func (res *DeleteRecordingsResponse) SetStatus(s int) {
	res.XMLResponse.SetStatus(s)
}

// UpdateRecordingsResponse indicates if the update was successful
// in the attribute updated. Might be different from Returncode.
// I guess.
type UpdateRecordingsResponse struct {
	*XMLResponse
	Updated bool `xml:"updated"`
}

// UnmarshalUpdateRecordingsResponse decodes the XML data
func UnmarshalUpdateRecordingsResponse(
	data []byte,
) (*UpdateRecordingsResponse, error) {
	res := &UpdateRecordingsResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal UpdateRecordingsResponse to XML
func (res *UpdateRecordingsResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// Header returns the HTTP response headers
func (res *UpdateRecordingsResponse) Header() http.Header {
	return res.XMLResponse.Header()
}

// SetHeader sets the HTTP response headers
func (res *UpdateRecordingsResponse) SetHeader(h http.Header) {
	res.XMLResponse.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *UpdateRecordingsResponse) Status() int {
	return res.XMLResponse.Status()
}

// SetStatus sets the HTTP response status code
func (res *UpdateRecordingsResponse) SetStatus(s int) {
	res.XMLResponse.SetStatus(s)
}

// GetDefaultConfigXMLResponse has the raw config xml data
type GetDefaultConfigXMLResponse struct {
	Config []byte

	header http.Header
	status int
}

// UnmarshalGetDefaultConfigXMLResponse creates a new response
// from the data.
func UnmarshalGetDefaultConfigXMLResponse(
	data []byte,
) (*GetDefaultConfigXMLResponse, error) {
	return &GetDefaultConfigXMLResponse{
		Config: data,
	}, nil
}

// Marshal GetDefaultConfigXMLResponse encodes the response
// body which is just the data.
func (res *GetDefaultConfigXMLResponse) Marshal() ([]byte, error) {
	if res.Config == nil {
		return nil, fmt.Errorf("no config is set in response")
	}
	return res.Config, nil
}

// Header returns the HTTP response headers
func (res *GetDefaultConfigXMLResponse) Header() http.Header {
	return res.Header()
}

// SetHeader sets the HTTP response headers
func (res *GetDefaultConfigXMLResponse) SetHeader(h http.Header) {
	res.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *GetDefaultConfigXMLResponse) Status() int {
	return res.Status()
}

// SetStatus sets the HTTP response status code
func (res *GetDefaultConfigXMLResponse) SetStatus(s int) {
	res.SetStatus(s)
}

// SetConfigXMLResponse encodes the result of setting the config
type SetConfigXMLResponse struct {
	*XMLResponse
	Token string `xml:"token"`
}

// UnmarshalSetConfigXMLResponse decodes the XML data
func UnmarshalSetConfigXMLResponse(
	data []byte,
) (*SetConfigXMLResponse, error) {
	res := &SetConfigXMLResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal encodes a SetConfigXMLResponse as XML
func (res *SetConfigXMLResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// Header returns the HTTP response headers
func (res *SetConfigXMLResponse) Header() http.Header {
	return res.XMLResponse.Header()
}

// SetHeader sets the HTTP response headers
func (res *SetConfigXMLResponse) SetHeader(h http.Header) {
	res.XMLResponse.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *SetConfigXMLResponse) Status() int {
	return res.XMLResponse.Status()
}

// SetStatus sets the HTTP response status code
func (res *SetConfigXMLResponse) SetStatus(s int) {
	res.XMLResponse.SetStatus(s)
}

// JSONResponse encapsulates a json reponse
type JSONResponse struct {
	Response interface{} `json:"response"`
}

// GetRecordingTextTracksResponse lists all tracks
type GetRecordingTextTracksResponse struct {
	Returncode string       `json:"returncode"`
	MessageKey string       `json:"messageKey,omitempty"`
	Message    string       `json:"message,omitempty"`
	Tracks     []*TextTrack `json:"tracks"`

	header http.Header
	status int
}

// UnmarshalGetRecordingTextTracksResponse decodes the json
func UnmarshalGetRecordingTextTracksResponse(
	data []byte,
) (*GetRecordingTextTracksResponse, error) {
	res := &JSONResponse{
		Response: &GetRecordingTextTracksResponse{},
	}
	err := json.Unmarshal(data, res)
	return res.Response.(*GetRecordingTextTracksResponse), err
}

// Marshal GetRecordingTextTracksResponse to JSON
func (res *GetRecordingTextTracksResponse) Marshal() ([]byte, error) {
	wrap := &JSONResponse{Response: res}
	return json.Marshal(wrap)
}

// Header returns the HTTP response headers
func (res *GetRecordingTextTracksResponse) Header() http.Header {
	return res.header
}

// SetHeader sets the HTTP response header
func (res *GetRecordingTextTracksResponse) SetHeader(h http.Header) {
	res.header = h
}

// Status returns the HTTP response status code
func (res *GetRecordingTextTracksResponse) Status() int {
	return res.status
}

// SetStatus sets the HTTP response status code
func (res *GetRecordingTextTracksResponse) SetStatus(s int) {
	res.status = s
}

// PutRecordingTextTrackResponse is the response when uploading
// a text track. Response is in JSON.
type PutRecordingTextTrackResponse struct {
	Returncode string `json:"returncode"`
	MessageKey string `json:"messageKey,omitempty"`
	Message    string `json:"message,omitempty"`
	RecordID   string `json:"recordId,omitempty"`

	header http.Header
	status int
}

// UnmarshalPutRecordingTextTrackResponse decodes the json response
func UnmarshalPutRecordingTextTrackResponse(
	data []byte,
) (*PutRecordingTextTrackResponse, error) {
	res := &JSONResponse{
		Response: &PutRecordingTextTrackResponse{},
	}
	err := json.Unmarshal(data, res)
	return res.Response.(*PutRecordingTextTrackResponse), err
}

// Marshal a PutRecordingTextTrackResponse to JSON
func (res *PutRecordingTextTrackResponse) Marshal() ([]byte, error) {
	wrap := &JSONResponse{Response: res}
	return json.Marshal(wrap)
}

// Header returns the HTTP response headers
func (res *PutRecordingTextTrackResponse) Header() http.Header {
	return res.header
}

// SetHeader sets the HTTP response header
func (res *PutRecordingTextTrackResponse) SetHeader(h http.Header) {
	res.header = h
}

// Status returns the HTTP response status code
func (res *PutRecordingTextTrackResponse) Status() int {
	return res.status
}

// SetStatus sets the HTTP response status code
func (res *PutRecordingTextTrackResponse) SetStatus(s int) {
	res.status = s
}

// Breakout info
type Breakout struct {
	XMLName         xml.Name `xml:"breakout"`
	ParentMeetingID string   `xml:"parentMeetingID"`
	Sequence        int      `xml:"sequence"`
	FreeJoin        bool     `xml:"freeJoin"`
}

// Attendee of a meeting
type Attendee struct {
	XMLName         xml.Name `xml:"attendee"`
	UserID          string   `xml:"userID"`
	InternalUserID  string   `xml:"internalUserID,omit"`
	FullName        string   `xml:"fullName"`
	Role            string   `xml:"role"`
	IsPresenter     bool     `xml:"isPresenter"`
	IsListeningOnly bool     `xml:"isListeningOnly"`
	HasJoinedVoice  bool     `xml:"hasJoinedVoice"`
	HasVideo        bool     `xml:"hasVideo"`
	ClientType      string   `xml:"clientType"`
}

// Meeting information
type Meeting struct {
	XMLName               xml.Name  `xml:"meeting"`
	MeetingName           string    `xml:"meetingName"`
	MeetingID             string    `xml:"meetingID"`
	InternalMeetingID     string    `xml:"internalMeetingID"`
	CreateTime            Timestamp `xml:"createTime"`
	CreateDate            string    `xml:"createDate"`
	VoiceBridge           string    `xml:"voiceBridge"`
	DialNumber            string    `xml:"dialNumber"`
	AttendeePW            string    `xml:"attendeePW"`
	ModeratorPW           string    `xml:"moderatorPW"`
	Running               bool      `xml:"running"`
	Duration              int       `xml:"duration"`
	Recording             bool      `xml:"recording"`
	HasBeenForciblyEnded  bool      `xml:"hasBeenForciblyEnded"`
	StartTime             Timestamp `xml:"startTime"`
	EndTime               Timestamp `xml:"endTime"`
	ParticipantCount      int       `xml:"participantCount"`
	ListenerCount         int       `xml:"listenerCount"`
	VoiceParticipantCount int       `xml:"voiceParticipantCount"`
	VideoCount            int       `xml:"videoCount"`
	MaxUsers              int       `xml:"maxUsers"`
	ModeratorCount        int       `xml:"moderatorCount"`
	IsBreakout            bool      `xml:"isBreakout"`

	Metadata Metadata `xml:"metadata"`

	Attendees     []*Attendee `xml:"attendees>attendee"`
	BreakoutRooms []string    `xml:"breakoutRooms>breakout"`
	Breakout      *Breakout   `xml:"breakout"`
}

func (m *Meeting) String() string {
	return fmt.Sprintf(
		"[Meeting id: %v, pc: %v, mc: %v, running: %v]",
		m.MeetingID, m.ParticipantCount, m.ModeratorCount, m.Running,
	)
}

// Update the meeting info with new data
func (m *Meeting) Update(update *Meeting) error {
	if m.MeetingID != update.MeetingID {
		return fmt.Errorf("meeting ids do not match for update")
	}
	if m.InternalMeetingID != update.InternalMeetingID {
		return fmt.Errorf("internal ids do not match for update")
	}
	/*

		if len(update.MeetingName) > 0 {
			m.MeetingName = update.MeetingName
		}
		if len(update.CreateDate) > 0 {
			m.CreateDate = update.CreateDate
		}
		if len(update.VoiceBridge) > 0 {
			m.VoiceBridge = update.VoiceBridge
		}
		if len(update.DialNumber) > 0 {
			m.DialNumber = update.DialNumber
		}
		if len(update.AttendeePW) > 0 {
			m.AttendeePW = update.AttendeePW
		}
		if len(update.ModeratorPW) > 0 {
			m.ModeratorPW = update.ModeratorPW
		}
		m.Running = update.Running
		m.Duration = update.Duration
		m.Recording = update.Recording
		m.HasBeenForciblyEnded = update.HasBeenForciblyEnded
		m.StartTime = update.StartTime
		m.EndTime = update.EndTime
		m.ParticipantCount = update.ParticipantCount
		m.ListenerCount = update.ListenerCount
		m.VoiceParticipantCount = update.VoiceParticipantCount
		m.VideoCount = update.VideoCount
		m.MaxUsers = update.MaxUsers
		m.ModeratorCount = update.ModeratorCount
		m.IsBreakout = update.IsBreakout
		m.Attendees = update.Attendees
		m.BreakoutRooms = update.BreakoutRooms
	*/

	*m = *update

	return nil
}

// Recording is a recorded bbb session
type Recording struct {
	XMLName           xml.Name  `xml:"recording"`
	RecordID          string    `xml:"recordID"`
	MeetingID         string    `xml:"meetingID"`
	InternalMeetingID string    `xml:"internalMeetingID"`
	Name              string    `xml:"name"`
	IsBreakout        bool      `xml:"isBreakout"`
	Published         bool      `xml:"published"`
	State             string    `xml:"state"`
	StartTime         Timestamp `xml:"startTime"`
	EndTime           Timestamp `xml:"endTime"`
	Participants      int       `xml:"participants"`
	Metadata          Metadata  `xml:"metadata"`
	Formats           []*Format `xml:"playback>format"`
}

// Format contains a link to the playable media
type Format struct {
	XMLName        xml.Name `xml:"format"`
	Type           string   `xml:"type"`
	URL            string   `xml:"url"`
	ProcessingTime int      `xml:"processingTime"` // No idea. The example is 7177.
	Length         int      `xml:"length"`
	Preview        *Preview `xml:"preview"`
}

// Preview contains a list of images
type Preview struct {
	XMLName xml.Name `xml:"preview"`
	Images  *Images  `xml:"images"`
}

// Images is a collection of Image
type Images struct {
	XMLName xml.Name `xml:"images"`
	All     []*Image `xml:"image"`
}

// Image is a preview image of the format
type Image struct {
	XMLName xml.Name `xml:"image"`
	Alt     string   `xml:"alt,attr"`
	Height  int      `xml:"height,attr"`
	Width   int      `xml:"width,attr"`
}

// TextTrack of a Recording
type TextTrack struct {
	Href   string `json:"href"`
	Kind   string `json:"kind"`
	Label  string `json:"label"`
	Source string `json:"source"`
}
