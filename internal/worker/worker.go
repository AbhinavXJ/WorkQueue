package worker

import (
	"WorkQueue/internal/task"
	"time"

	"fmt"
)

func Process_Task(task_to_execute task.Task) error {

	if task_to_execute.Payload == nil {
		return fmt.Errorf("payload is empty")
	}

	//Add your task type here , and perform the task under your switch case
	switch task_to_execute.Type {

	case "send_email":
		time.Sleep(2 * time.Second)
		fmt.Println("Sending email to ", task_to_execute.Payload["to"], " with subject ", task_to_execute.Payload["subject"])
		return nil
	case "resize_image":
		fmt.Println("Resizing image to x cordinate: ", task_to_execute.Payload["new_x"], " y cordinate: ", task_to_execute.Payload["new_y"])
		return nil
	case "generate_pdf":
		fmt.Println("Generating pdf...")
		return nil

	case "":
		return fmt.Errorf("task type is empty")
	default:
		return fmt.Errorf("unsupported task")

	}
}
