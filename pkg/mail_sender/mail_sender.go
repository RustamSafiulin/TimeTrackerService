package mail_sender

type Config struct {
	Server   string
	Port     int
	Emai     string
	Password string
}

type Request struct {
	From    string
	To      []string
	Subject string
	Body    string
}

func Send() {

}
