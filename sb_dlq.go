package golang_servicebus_cli

import (
	"context"
	"fmt"
	servicebus "github.com/Azure/azure-service-bus-go"
	"os"
	"time"
)

func GetDQLMessages(connectionString string, queueName string){
	ctx,cancel:=context.WithTimeout(context.Background(),40*time.Second)
	defer cancel()

	// Create a client to communicate with a Service Bus Namespace.
	ns, err := servicebus.NewNamespace(servicebus.NamespaceWithConnectionString(connectionString))
	if err != nil {
		fmt.Println(err)
		return
	}

	qm := ns.NewQueueManager()
		qe,err:=qm.Get(ctx,queueName)
	if err != nil {
		fmt.Println(err)
		return
	}

	q, err := ns.NewQueue(qe.Name)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		_ = q.Close(ctx)
	}()

	if err := q.Send(ctx, servicebus.NewMessageFromString("foo")); err != nil {
		fmt.Println(err)
		return
	}

	// Abandon the message 10 times simulating attempting to process the message 10 times. After the 10th time, the
	// message will be placed in the Deadletter Queue.
	for count := 0; count < 10; count++ {
		err = q.ReceiveOne(ctx, servicebus.HandlerFunc(func(ctx context.Context, msg *servicebus.Message) error {
			fmt.Printf("count: %d\n", count+1)
			return msg.Abandon(ctx)
		}))
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// receive one from the queue's deadletter queue. It should be the foo message.
	qdl := q.NewDeadLetter()

	if err := qdl.ReceiveOne(ctx, MessagePrinter{}); err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		_ = qdl.Close(ctx)
	}()

}
