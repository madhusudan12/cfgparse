package cfgparse


func main() {
	cfg := New()
	err := cfg.ReadFile("config.ini")
	if err!=nil {
		panic(err)
	}



}
