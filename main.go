package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"strconv"

	"github.com/disintegration/imaging"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func findMinDimensions(images []image.Image) (int, int) {
	minWidth, minHeight := images[0].Bounds().Dx(), images[0].Bounds().Dy()

	for _, img := range images {
		width, height := img.Bounds().Dx(), img.Bounds().Dy()
		if width < minWidth {
			minWidth = width
		}
		if height < minHeight {
			minHeight = height
		}
	}

	return minWidth, minHeight
}

func interlaceImages(images []image.Image, stripWidth int) (image.Image, error) {
	width := images[0].Bounds().Dx()
	height := images[0].Bounds().Dy()

	interlaced := image.NewNRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		imgIndex := (x / stripWidth) % len(images)
		for y := 0; y < height; y++ {
			interlaced.Set(x, y, images[imgIndex].At(x, y))
		}
	}

	return interlaced, nil
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		writeJSONError(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Parse LPI from form data (optional, defaults to 10)
	lpiStr := r.FormValue("lpi")
	lpi, err := strconv.Atoi(lpiStr)
	if err != nil || lpi <= 0 {
		lpi = 50 // Default LPI hai
	}

	dpiStr := r.FormValue("dpi")
	dpi, err := strconv.Atoi(dpiStr)
	if err != nil || dpi <= 0 {
		dpi = 96 // This is Default DPI
	}

	stripWidth := dpi / lpi
	fmt.Printf("Calculated strip width: %d pixels per strip\n", stripWidth)

	files := r.MultipartForm.File["images"]
	var loadedImages []image.Image

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			fmt.Println(err)
			writeJSONError(w, "Unable to open uploaded file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		img, err := imaging.Decode(file)
		if err != nil {
			fmt.Println(err)
			writeJSONError(w, "Failed to decode image", http.StatusBadRequest)
			return
		}
		loadedImages = append(loadedImages, img)
	}

	minWidth, minHeight := findMinDimensions(loadedImages)

	var resizedImages []image.Image
	for _, img := range loadedImages {
		resizedImg := imaging.Resize(img, minWidth, minHeight, imaging.Lanczos)
		resizedImages = append(resizedImages, resizedImg)
	}

	interlacedImg, err := interlaceImages(resizedImages, stripWidth)
	if err != nil {
		fmt.Println(err)
		writeJSONError(w, "Failed to interlace images", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	if err := png.Encode(w, interlacedImg); err != nil {
		fmt.Println(err)
		writeJSONError(w, "Failed to encode interlaced image", http.StatusInternalServerError)
		return
	}
}

func main() {
	router := mux.NewRouter()
	router.Use(cors.Default().Handler)
	router.HandleFunc("/upload", uploadHandler).Methods("POST")
	fmt.Println("Server started at :8081")
	http.ListenAndServe(":8081", router)
}
