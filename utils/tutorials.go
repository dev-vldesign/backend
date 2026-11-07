package utils

func GetVideoDuration(filename string) (float64, error) {
	// Temporarily return a default duration since ffmpeg is not available
	// TODO: Re-enable ffmpeg functionality when needed
	return 300.0, nil // Return 5 minutes as default duration

	// ## THIS IS FOR USING ffmpeg FOR WINDOWS ONLY
	// cmd := exec.Command("C:\\ffmpeg\\bin\\ffprobe.exe", "-v", "error", "-show_entries",
	// "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filename)

	// ## THIS SHOULD BE USED ON THE UBUNTU SERVER

	// cmd := exec.Command("ffprobe", "-v", "error", "-show_entries",
	// "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filename)

	// output, err := cmd.Output()
	// if err != nil {
	// 	return 0, err
	// }

	// duration, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	// if err != nil {
	// 	return 0, err
	// }
	// return duration / (1000000000), nil // duration in minutes
}
