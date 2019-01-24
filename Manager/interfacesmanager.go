package Data

func IMInit() error {
	err := InitDevicesData()
	return err
}

func IMStart(ch *chan error) {
	var err error
	//TODO
	for true{
		if false{
			break
		}
	}
	defer func(e error) {
		*ch <- e
	}(err)
}

func IMShutDown() error{
	return nil
}

