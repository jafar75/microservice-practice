package main

import (
	"fmt"
	"context"

	"github.com/jafar75/microservice-practice/application"

)

func main() {
	app := application.New();

	err := app.Start(context.TODO());

	if err != nil {
		fmt.Println("failed to start app with err: ", err);
	}
}

