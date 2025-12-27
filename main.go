package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

func main() {

	logFile, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Log dosyası oluşturulamadı:", err)
	}
	defer logFile.Close()

	// Hem ekrana hem dosyaya yazmak için
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger := log.New(multiWriter, "", log.LstdFlags)

	// Tor Proxy 9150 veya 9050
	proxyAddr := "127.0.0.1:9150" 
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		logger.Fatalf("[CRITICAL] Proxy bağlantı hatası: %v", err)
	}

	transport := &http.Transport{Dial: dialer.Dial}
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second * 30,
	}

	// 	Tor ağ kontrolü
	logger.Println("[CHECK] Tor ağ bağlantısı kontrol ediliyor...")
	checkResp, err := client.Get("http://check.torproject.org")
	if err != nil || checkResp.StatusCode != 200 {
		logger.Fatalf("[ERROR] Tor ağına bağlanılamadı! Lütfen Tor Browser'ın açık olduğundan emin olun.")
	}
	logger.Println("[SUCCESS] Tor ağı aktif. Tarama başlıyor...")


	outputDir := "outputs"
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.Mkdir(outputDir, 0755)
	}

	targetsFile, err := os.Open("targets.yaml")
	if err != nil {
		logger.Fatalf("[ERROR] targets.yaml bulunamadı: %v", err)
	}
	defer targetsFile.Close()

	scanner := bufio.NewScanner(targetsFile)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url == "" || strings.HasPrefix(url, "#") {
			continue
		}

		logger.Printf("[INFO] Taranıyor: %s", url)

		resp, err := client.Get(url)
		if err != nil {
			logger.Printf("[ERR] Başarısız: %s | Hata: %v", url, err)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			logger.Printf("[ERR] Body okuma hatası: %s", url)
			continue
		}

		timestamp := time.Now().Format("20060102_150405")
		cleanUrl := strings.TrimPrefix(url, "http://")
		cleanUrl = strings.TrimPrefix(cleanUrl, "https://")
		cleanUrl = strings.ReplaceAll(cleanUrl, "/", "_")
		
		fileName := fmt.Sprintf("%s/%s_%s.html", outputDir, timestamp, cleanUrl)

		err = ioutil.WriteFile(fileName, body, 0644)
		if err != nil {
			logger.Printf("[ERR] Kayıt hatası: %v", err)
		} else {
			logger.Printf("[SUCCESS] Kaydedildi: %s", fileName)
		}
	}
	logger.Println("[FINISH] Tüm işlemler tamamlandı.")
}