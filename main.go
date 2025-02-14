package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var (
	androidPackage         string
	androidFingerprint    string
	iosTeamID             string
	iosBundleID           string
	iosAppStoreID         string
	androidStoreURL       string
	iosStoreURL           string
)

func init() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	// Load configuration from environment variables
	androidPackage = getEnvOrDefault("ANDROID_PACKAGE_NAME", "com.mhmd.wasl")
	androidFingerprint = getEnvOrDefault("ANDROID_SHA256_FINGERPRINT", "FD:F7:95:1B:2E:25:BF:4C:19:6F:48:91:A1:04:8A:82:71:ED:08:62:30:E5:93:5B:E9:2D:09:9A:4E:48:62:AF")
	iosTeamID = getEnvOrDefault("IOS_TEAM_ID", "YOUR_TEAM_ID")
	iosBundleID = getEnvOrDefault("IOS_BUNDLE_ID", "com.mhmd.wasl")
	iosAppStoreID = getEnvOrDefault("IOS_APP_STORE_ID", "YOUR_APP_ID")
	androidStoreURL = getEnvOrDefault("ANDROID_STORE_URL", "https://play.google.com/store/apps/details?id="+androidPackage)
	iosStoreURL = getEnvOrDefault("IOS_STORE_URL", "https://apps.apple.com/app/"+iosAppStoreID)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// AssetLinks represents the Android assetlinks.json structure
type AssetLinks []struct {
	Relation []string `json:"relation"`
	Target   struct {
		Namespace             string   `json:"namespace"`
		PackageName          string   `json:"package_name"`
		Sha256CertFingerprints []string `json:"sha256_cert_fingerprints"`
	} `json:"target"`
}

// AppleAppSiteAssociation represents the iOS apple-app-site-association structure
type AppleAppSiteAssociation struct {
	AppLinks struct {
		Apps    []string `json:"apps"`
		Details []struct {
			AppID string   `json:"appID"`
			Paths []string `json:"paths"`
		} `json:"details"`
	} `json:"applinks"`
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
}

func main() {
	// Add CORS middleware
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Route the request
		switch {
		case strings.HasPrefix(r.URL.Path, "/.well-known/assetlinks.json"):
			handleAssetLinks(w, r)
		case strings.HasPrefix(r.URL.Path, "/.well-known/apple-app-site-association"):
			handleAppleAppSiteAssociation(w, r)
		case strings.HasPrefix(r.URL.Path, "/store/") || 
			 strings.HasPrefix(r.URL.Path, "/item/") || 
			 strings.HasPrefix(r.URL.Path, "/combo/"):
			handleRedirect(w, r)
		default:
			http.NotFound(w, r)
		}
	})

	// Get port from environment variable
	port := getEnvOrDefault("PORT", "8080")

	fmt.Printf("Server running on port :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

func handleAssetLinks(w http.ResponseWriter, r *http.Request) {
	assetLinks := AssetLinks{
		{
			Relation: []string{"delegate_permission/common.handle_all_urls"},
			Target: struct {
				Namespace             string   `json:"namespace"`
				PackageName          string   `json:"package_name"`
				Sha256CertFingerprints []string `json:"sha256_cert_fingerprints"`
			}{
				Namespace:    "android_app",
				PackageName: androidPackage,
				Sha256CertFingerprints: []string{
					androidFingerprint,
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assetLinks)
}

func handleAppleAppSiteAssociation(w http.ResponseWriter, r *http.Request) {
	appSiteAssociation := AppleAppSiteAssociation{
		AppLinks: struct {
			Apps    []string `json:"apps"`
			Details []struct {
				AppID string   `json:"appID"`
				Paths []string `json:"paths"`
			} `json:"details"`
		}{
			Apps: []string{},
			Details: []struct {
				AppID string   `json:"appID"`
				Paths []string `json:"paths"`
			}{
				{
					AppID: fmt.Sprintf("%s.%s", iosTeamID, iosBundleID),
					Paths: []string{
						"/store/*",
						"/item/*",
						"/combo/*",
					},
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(appSiteAssociation)
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	userAgent := strings.ToLower(r.UserAgent())

	// Create the deep link URL
	var deepLink string
	if strings.Contains(userAgent, "android") {
		// Android device
		deepLink = fmt.Sprintf("intent://%s#Intent;scheme=wasl;package=%s;end", path, androidPackage)
	} else if strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipad") || strings.Contains(userAgent, "ipod") {
		// iOS device
		deepLink = fmt.Sprintf("wasl://%s", path)
	} else {
		// Desktop or unknown device - show a landing page
		renderLandingPage(w, path)
		return
	}

	// Redirect to the appropriate deep link
	w.Header().Set("Location", deepLink)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func renderLandingPage(w http.ResponseWriter, path string) {
	html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html dir="rtl">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>وصل - فتح التطبيق</title>
			<style>
				body {
					font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
					margin: 0;
					padding: 20px;
					display: flex;
					flex-direction: column;
					align-items: center;
					justify-content: center;
					min-height: 100vh;
					background: linear-gradient(135deg, #6B73FF 0%, #000DFF 100%%);
					color: white;
					text-align: center;
				}
				.container {
					background: rgba(255, 255, 255, 0.1);
					padding: 40px;
					border-radius: 20px;
					backdrop-filter: blur(10px);
					max-width: 500px;
					width: 90%%;
				}
				h1 {
					margin-bottom: 20px;
				}
				.store-buttons {
					display: flex;
					flex-direction: column;
					gap: 15px;
					margin-top: 30px;
				}
				.store-button {
					padding: 15px 30px;
					border-radius: 10px;
					background: white;
					color: #000DFF;
					text-decoration: none;
					font-weight: bold;
					transition: transform 0.2s;
				}
				.store-button:hover {
					transform: scale(1.05);
				}
			</style>
		</head>
		<body>
			<div class="container">
				<h1>وصل</h1>
				<p>للحصول على أفضل تجربة، قم بتحميل تطبيق وصل</p>
				<div class="store-buttons">
					<a href="%s" class="store-button">
						تحميل من Google Play
					</a>
					<a href="%s" class="store-button">
						تحميل من App Store
					</a>
				</div>
			</div>
		</body>
		</html>
	`, androidStoreURL, iosStoreURL)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}
