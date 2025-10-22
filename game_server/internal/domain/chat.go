package domain

type Chat struct {
	From string `json:"from"`
	Id   string `json:"id"`
	Text string `json:"text"`
}

func (c *Chat) Clone() *Chat {
	return &Chat{
		From: c.From,
		Id:   c.Id,
		Text: c.Text,
	}
}
