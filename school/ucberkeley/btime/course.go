package btime

// Course is the course info sent back as json.
type Course struct {
	Course struct {
		Units              string  `json:"units"`
		Description        string  `json:"description"`
		Title              string  `json:"title"`
		Prerequisites      string  `json:"prerequisites"`
		Abbreviation       string  `json:"abbreviation"`
		EnrolledPercentage float64 `json:"enrolled_percentage"`
		Department         string  `json:"department"`
		EnrolledMax        int     `json:"enrolled_max"`
		Waitlisted         int     `json:"waitlisted"`
		Enrolled           int     `json:"enrolled"`
		GradeAverage       float64 `json:"grade_average"`
		CourseNumber       string  `json:"course_number"`
		ID                 int     `json:"id"`
		LetterAverage      string  `json:"letter_average"`
	} `json:"course"`
	Marketplace struct {
	} `json:"marketplace"`
	LastEnrollmentUpdate string   `json:"last_enrollment_update"`
	Requirements         []string `json:"requirements"`
	Favorited            bool     `json:"favorited"`
	Ongoing              bool     `json:"ongoing"`
	CoverPhoto           string   `json:"cover_photo"`
	Sections             []struct {
		Kind          string `json:"kind"`
		LocationName  string `json:"location_name"`
		Waitlisted    int    `json:"waitlisted"`
		FinalEnd      string `json:"final_end"`
		StartTime     string `json:"start_time"`
		SectionNumber string `json:"section_number"`
		FinalStart    string `json:"final_start"`
		WordDays      string `json:"word_days"`
		Ccn           string `json:"ccn"`
		EnrolledMax   int    `json:"enrolled_max"`
		EndTime       string `json:"end_time"`
		FinalDay      string `json:"final_day"`
		Enrolled      int    `json:"enrolled"`
		Instructor    string `json:"instructor"`
		ID            int    `json:"id"`
	} `json:"sections"`
	OngoingSections []struct {
		Kind          string `json:"kind"`
		LocationName  string `json:"location_name"`
		Waitlisted    int    `json:"waitlisted"`
		FinalEnd      string `json:"final_end"`
		StartTime     string `json:"start_time"`
		SectionNumber string `json:"section_number"`
		FinalStart    string `json:"final_start"`
		WordDays      string `json:"word_days"`
		Ccn           string `json:"ccn"`
		EnrolledMax   int    `json:"enrolled_max"`
		EndTime       string `json:"end_time"`
		FinalDay      string `json:"final_day"`
		Enrolled      int    `json:"enrolled"`
		Instructor    string `json:"instructor"`
		ID            int    `json:"id"`
	} `json:"ongoing_sections"`
}
