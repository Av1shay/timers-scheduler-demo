package server

type SetTimerReq struct {
	Hours   int    `json:"hours" validate:"gte=0"`
	Minutes int    `json:"minutes" validate:"gte=0"`
	Seconds int    `json:"seconds" validate:"gte=0"`
	URL     string `json:"url" validate:"empty=false&format=url"`
}

type SetTimerResp struct {
	ID int `json:"id"`
}

type GetTimerResp struct {
	ID       int   `json:"id"`
	TimeLeft int64 `json:"time_left"`
}
