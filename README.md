# leopard
golang websocket

## main.go
   l := engine.New()
#### the way 1

     
	   l.AddSpace(&controller.Index{})

	   l.OnConnect(func(client iFace.IClient) {
		
	   })

	   l.OnPing(func(client iFace.IClient) {
		     l.Logger().Infof("%s ping %s", strings.Repeat("-", 30), strings.Repeat("-", 30))
	   })

	   l.OnMessage(func(client iFace.IClient, request []byte) {
		
	   })

	   l.OnDisconnect(func(client iFace.IClient) {
		
	   })
#### the way 2
     l.WithOptions(engine.Options{
		     Config: &iFace.Config{
			      WriteWait:         10 * time.Second,
			      PongWait:          60 * time.Second,
			      PingPeriod:        (60 * time.Second * 9) / 10,
			      MaxMessageSize:    512,
			      MessageBufferSize: 256,
			      ReadBufferSize:    1024,
			      WriteBufferSize:   1024,
		      },
		      HandleFunc: engine.HandleFunc{
			        Error: nil,
			        Message: func(client iFace.IClient, request []byte) {
				
			        },
			        Connect: func(client iFace.IClient) {
				
			        },
			        Disconnect: func(client iFace.IClient) {
				
			        },
		       },
	      })
        
  l.Run()
