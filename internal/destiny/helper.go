package destiny

type Helper struct {
	client *Client
}

func NewHelper(client *Client) *Helper {
	return &Helper{client: client}
}
