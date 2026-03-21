package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// Ensure this matches your .env PORT!
const baseURL = "http://localhost:9090/api/v1"

func main() {
	// 1. Check if the user has a real MP4 file they want to test with
	// If not, we generate a tiny text file pretending to be an mp4
	fileName := "test_movie.mp4"
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		fmt.Println("⚠️  No real 'test_movie.mp4' found in the root directory. Generating a tiny fake video file just for testing the API logic...")
		os.WriteFile(fileName, []byte("fake video content..."), 0644)
		defer os.Remove(fileName) // Cleanup the fake file at the end
	} else {
		fmt.Println("🎥 Found real 'test_movie.mp4'! Transcoding will take some time depending on its size.")
	}

	// 2. Register/Login to get JWT Token
	fmt.Println("👤 Registering temporary user...")
	authBody := []byte(`{"email":"intern@streaming.com","username":"intern_test","password":"password123"}`)
	
	resp, err := http.Post(baseURL+"/auth/register", "application/json", bytes.NewBuffer(authBody))
	if err != nil {
		fmt.Println("❌ Failed to reach server. Make sure 'go run ./cmd/api/main.go' is running in another terminal!")
		os.Exit(1)
	}
	
	// If already registered (400 or 409), try logging in instead
	if resp.StatusCode >= 400 {
		resp, _ = http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(authBody))
	}

	var authResult struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResult); err != nil {
		fmt.Println("❌ Failed to parse JWT token. Did the login endpoint error out?")
		os.Exit(1)
	}
	resp.Body.Close()

	if authResult.Token == "" {
		fmt.Println("❌ Received empty token. Registration/Login failed.")
		os.Exit(1)
	}
	fmt.Println("✅ Successfully acquired JWT Token!")

	// 3. Upload Video using multipart form
	fmt.Println("🎬 Uploading video file to POST /api/v1/videos ...")
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	
	writer.WriteField("title", "My First Test Video")
	writer.WriteField("description", "Testing the Golang HLS streaming backend!")

	part, _ := writer.CreateFormFile("video", fileName)
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("❌ Failed to open video file.")
		os.Exit(1)
	}
	io.Copy(part, file)
	file.Close()
	writer.Close()

	req, _ := http.NewRequest("POST", baseURL+"/videos", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+authResult.Token) // Pass the JWT!

	respUpload, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("❌ Failed to upload video: " + err.Error())
		os.Exit(1)
	}
	
	respBytes, _ := io.ReadAll(respUpload.Body)
	respUpload.Body.Close()
	
	fmt.Printf("📬 Server Response (HTTP %d):\n%s\n", respUpload.StatusCode, string(respBytes))
	
	if respUpload.StatusCode == 201 {
		fmt.Println("🎉 SUCCESS! Video upload HTTP request accepted.")
		fmt.Println("⚙️  Check your 'go run' server logs to watch the background FFmpeg worker!")
	} else {
		fmt.Println("❌ FAILED to upload.")
	}
}
