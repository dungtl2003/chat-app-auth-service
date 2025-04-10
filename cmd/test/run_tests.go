package main

import (
	"database/sql"
	"dungtl2003/chat-app-auth-service/internal/model"
	"dungtl2003/chat-app-auth-service/internal/password"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"sync"

	_ "github.com/lib/pq" // postgresql driver support
)

const (
	USER_DATA_FILE = "chat_users.json"
)

func runSetup(client *sql.DB, passwordManager password.PasswordManager) error {
	fmt.Println("Setting up...")
	err := cleanDb(client)
	if err != nil {
		return err
	}

	projectDir, err := os.Getwd()
	if err != nil {
		return err
	}

	userDataFilePath := path.Join(projectDir, "tests", "data", USER_DATA_FILE)
	userData, err := os.ReadFile(userDataFilePath)
	if err != nil {
		return err
	}
	var users []model.ChatUser
	err = json.Unmarshal(userData, &users)
	if err != nil {
		fmt.Println(err)
		return err
	}

	tx, err := client.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	fmt.Println("Creating temporary users")
	for _, user := range users {
		password, err := passwordManager.Hash(user.Password)
		if err != nil {
			return fmt.Errorf("Error hashing password: %v\n", err)
		}

		_, err = tx.Exec(`INSERT INTO chat_user.chat_user (
            id, email, username, password, role
        ) VALUES (
            $1, $2, $3, $4, $5
        );`, user.Id, user.Email, user.Username, password, user.Role)
		if err != nil {
			return err
		}

		for _, device := range user.Devices {
			_, err = tx.Exec(`INSERT INTO chat_user.device (
                id, device_name, device_type, os, status, user_id
            ) VALUES (
                $1, $2, $3, $4, $5, $6
            );`, device.Id, device.DeviceName, device.DeviceType, device.Os, device.Status, user.Id)
			if err != nil {
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func runTearDown(client *sql.DB) error {
	fmt.Println("Tearing down...")
	err := cleanDb(client)
	if err != nil {
		return err
	}

	return nil
}

func runTests() error {
	isJson := flag.Bool("json", false, "Output test results in json format")
	flag.Parse()

	if *isJson {
		fmt.Println("Running tests with json format")
		cmd := exec.Command("go", "test", "-json", "-v", "./...")
		outputFile, has := os.LookupEnv("TEST_OUT")
		if !has {
			outputFile = "test_output.json"
			fmt.Printf("Warning: `TEST_OUT` is not set, output json to default file: %s\n", outputFile)
		}

		var err error
		teeCmd := exec.Command("tee", outputFile)
		teeCmd.Stdin, err = cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("Error creating pipe: %v\n", err)
		}
		teeCmd.Stdout = os.Stdout // Output to terminal as well

		err = teeCmd.Start()
		if err != nil {
			return fmt.Errorf("Error starting tee: %v\n", err)
		}

		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("Error running go test: %v\n", err)
		}

		err = teeCmd.Wait()
		if err != nil {
			return fmt.Errorf("Error waiting for tee command: %v\n", err)
		}
	} else {
		fmt.Println("Running tests")
		cmd := exec.Command("go", "test", "-v", "./...")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("Error running go test: %v\n", err)
		}
	}

	return nil
}

func main() {
	costStr, has := os.LookupEnv("COST")
	cost := 12
	if has {
		var err error
		cost, err = strconv.Atoi(costStr)
		if err != nil {
			fmt.Printf("Invalid cost = %s\n", costStr)
			os.Exit(1)
		}
	}

	passwordManager, err := password.NewBcryptPasswordManager(cost)
	if err != nil {
		fmt.Printf("Error creating password manager: %v\n", err)
		os.Exit(1)
	}

	adminDbURL, has := os.LookupEnv("ADMIN_DATABASE_URL")
	if !has {
		fmt.Println("Missing ADMIN_DATABASE_URL")
		os.Exit(1)
	}

	dbLog, has := os.LookupEnv("DB_LOG")
	if has {
		defer func() {
			err = saveLog("chat-app-db-service", dbLog)
			if err != nil {
				fmt.Printf("saveLog(): %v\n", err)
			}
		}()
	}

	userLog, has := os.LookupEnv("USER_SERVICE_LOG")
	if has {
		defer func() {
			err = saveLog("chat-app-user-service", userLog)
			if err != nil {
				fmt.Printf("saveLog(): %v\n", err)
			}
		}()
	}

	snowflakeLog, has := os.LookupEnv("SNOWFLAKE_SERVICE_LOG")
	if has {
		defer func() {
			err = saveLog("chat-app-snowflake-service", snowflakeLog)
			if err != nil {
				fmt.Printf("saveLog(): %v\n", err)
			}
		}()
	}

	// dbLog, has := os.LookupEnv("DB_LOG")
	// if has {
	// 	fmt.Println("Start collecting database service log")
	// 	lc, err := StartLogCapture("chat-app-db-service", dbLog)
	// 	if err != nil {
	// 		fmt.Printf("Failed capturing database service log: %v", err)
	// 		os.Exit(1)
	// 	}
	// 	defer func() {
	// 		err = lc.Stop()
	// 		if err != nil {
	// 			fmt.Printf("Stop(): %v", err)
	// 		}
	// 	}()
	// }
	//
	// userLog, has := os.LookupEnv("USER_SERVICE_LOG")
	// if has {
	// 	fmt.Println("Start collecting user service log")
	// 	lc, err := StartLogCapture("chat-app-user-service", userLog)
	// 	if err != nil {
	// 		fmt.Printf("Failed capturing user service log: %v", err)
	// 		os.Exit(1)
	// 	}
	// 	defer func() {
	// 		err = lc.Stop()
	// 		if err != nil {
	// 			fmt.Printf("Stop(): %v", err)
	// 		}
	// 	}()
	// }
	//
	// snowflakeLog, has := os.LookupEnv("SNOWFLAKE_SERVICE_LOG")
	// if has {
	// 	fmt.Println("Start collecting snowflake service log")
	// 	lc, err := StartLogCapture("chat-app-snowflake-service", snowflakeLog)
	// 	if err != nil {
	// 		fmt.Printf("Failed capturing snowflake service log: %v", err)
	// 		os.Exit(1)
	// 	}
	// 	defer func() {
	// 		err = lc.Stop()
	// 		if err != nil {
	// 			fmt.Printf("Stop(): %v", err)
	// 		}
	// 	}()
	// }

	client, err := sql.Open("postgres", adminDbURL)
	if err != nil {
		fmt.Println("Error opening database connection:", err)
		os.Exit(1)
	}
	defer func() {
		fmt.Println("Closing database connection")
		client.Close()
	}()

	err = runSetup(client, passwordManager)
	if err != nil {
		fmt.Println("Error running setup:", err)
		os.Exit(1)
	}
	// fmt.Println("Press any key to continue")
	// input := bufio.NewScanner(os.Stdin)
	// input.Scan()

	err = runTests()
	if err != nil {
		fmt.Println("Error running tests:", err)
		os.Exit(1)
	}

	err = runTearDown(client)
	if err != nil {
		fmt.Println("Error running teardown:", err)
		os.Exit(1)
	}
}

func saveLog(serviceName string, logFilePath string) error {
	file, err := os.Create(logFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	cmd := exec.Command("docker", "logs", serviceName)
	cmd.Stdout = file
	cmd.Stderr = file
	return cmd.Run()
}

func cleanDb(client *sql.DB) error {
	_, err := client.Exec(`DELETE FROM chat_user.chat_user;`) // this will also delete conversations, participants, and messages (cascade)
	if err != nil {
		return err
	}

	return nil
}

type LogCapture struct {
	cmd  *exec.Cmd
	file *os.File
	done chan struct{}
	wg   sync.WaitGroup
	err  error
}

func StartLogCapture(serviceName string, logFilePath string) (*LogCapture, error) {
	// Open file for writing logs
	fmt.Printf("Capturing %s and save in %s\n", serviceName, logFilePath)
	file, err := os.Create(logFilePath)
	if err != nil {
		return nil, err
	}

	// Start docker logs command with --follow
	cmd := exec.Command("docker", "logs", serviceName, "--follow")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		file.Close()
		return nil, err
	}

	lc := &LogCapture{
		cmd:  cmd,
		file: file,
		done: make(chan struct{}),
	}

	// Start copying logs to file in a goroutine
	lc.wg.Add(1)
	go func() {
		defer lc.wg.Done()
		_, lc.err = io.Copy(file, stdout)
	}()

	// Start the command
	if err := cmd.Start(); err != nil {
		file.Close()
		return nil, err
	}

	return lc, nil
}

func (lc *LogCapture) Stop() error {
	// Signal the command to stop
	lc.cmd.Process.Signal(os.Interrupt)
	close(lc.done)

	// Wait for the copy goroutine to finish
	lc.wg.Wait()

	// Close the file
	fileErr := lc.file.Close()

	// Wait for the command to exit and get any error
	cmdErr := lc.cmd.Wait()

	if lc.err != nil {
		fmt.Printf("lc")
		return lc.err
	}
	if fileErr != nil {
		fmt.Printf("file")
		return fileErr
	}
	return cmdErr
}
