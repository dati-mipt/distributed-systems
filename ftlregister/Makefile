test:
	go test -v -run TestFaultTolerantRegisterRealtime -count 50 -failfast
	go test -v -run TestFaultTolerantRegisterIncOrder -count 50 -failfast
	go test -v -run TestFaultTolerantRegisterDieHard -count 100 -failfast
