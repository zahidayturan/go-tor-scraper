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
	outputDir := "outputs"
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err := os.Mkdir(outputDir, 0755)
		if err != nil {
			log.Fatal("Outputs klasörü oluşturulamadı:", err)
		}
	}

	logFile, err := os.OpenFile(outputDir+"/logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Log dosyası oluşturulamadı:", err)
	}
	defer logFile.Close()

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger := log.New(multiWriter, "", log.LstdFlags)

	proxyAddr := "127.0.0.1:9150" 
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		logger.Fatalf("[CRITICAL] [CLOSE] Proxy sunucusuna bağlanılamadı: %v \n\n\n", err)
	}

	transport := &http.Transport{Dial: dialer.Dial}
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second * 60,
	}

	logger.Println("[CHECK] Tor ağ bağlantısı doğrulanıyor...")
	checkResp, err := client.Get("http://check.torproject.org")
	if err != nil || checkResp.StatusCode != 200 {
		logger.Fatalf("[ERROR] [CLOSE] Tor ağı aktif değil! Lütfen Tor Browser'ı açın ve portu (9150/9050) kontrol edin. \n\n\n")
	}
	logger.Println("[SUCCESS] Tor ağı bağlantısı sağlandı. İşlemler başlıyor.")

	targetsFile, err := os.Open("targets.yaml")
	if err != nil {
		logger.Fatalf("[ERROR] [CLOSE] targets.yaml dosyası bulunamadı: %v \n\n\n", err)
	}
	defer targetsFile.Close()

	totalSites := 0
	successCount := 0
	failCount := 0

	scanner := bufio.NewScanner(targetsFile)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		
		if url == "" || strings.HasPrefix(url, "#") || strings.HasPrefix(url, "---") {
			continue
		}

		totalSites++
		logger.Printf("[INFO] Taranıyor (%d): %s", totalSites, url)

		resp, err := client.Get(url)
		if err != nil {
			logger.Printf("[ERR] Siteye ulaşılamadı: %s | Hata: %v", url, err)
			failCount++
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			logger.Printf("[ERR] İçerik okuma hatası: %s", url)
			failCount++
			continue
		}

		timestamp := time.Now().Format("20060102_150405")

		cleanUrl := strings.TrimPrefix(url, "http://")
		cleanUrl = strings.TrimPrefix(cleanUrl, "https://")
		cleanUrl = strings.ReplaceAll(cleanUrl, "/", "_")
		cleanUrl = strings.ReplaceAll(cleanUrl, ":", "_")
		
		fileName := fmt.Sprintf("%s/%s_%s.html", outputDir, timestamp, cleanUrl)

		err = ioutil.WriteFile(fileName, body, 0644)
		if err != nil {
			logger.Printf("[ERR] Kayıt hatası (%s): %v", url, err)
			failCount++
		} else {
			logger.Printf("[SUCCESS] Veri kaydedildi: %s", fileName)
			successCount++
		}
	}

	logger.Printf("[REPORT] Toplam Taranan : %d | Başarılı : %d | Başarısız : %d", totalSites, successCount, failCount)
	logger.Println("[FINISH] Tüm işlemler tamamlandı.")
	logger.Println("[CLOSE] ---------------------------------------------\n\n\n")
}