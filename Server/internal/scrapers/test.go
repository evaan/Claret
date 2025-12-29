package scrapers

// func main() {
// 	jsession, err := GetJsession()
// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	jar, _ := cookiejar.New(nil)
// 	client := &http.Client{Jar: jar}
// 	baseUrl, _ := url.Parse("https://self-service.mun.ca")
// 	jar.SetCookies(baseUrl, []*http.Cookie{
// 		{Name: "JSESSIONID", Value: jsession, Path: "/StudentRegistrationSsb"},
// 	})

// 	semesters, err := GetSemesters()
// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	for _, semester := range semesters {
// 		crns, err := GetCRNs(client, jsession, semester.Code)
// 		if err != nil {
// 			log.Fatalln(err)
// 		}

// 		for i, crn := range crns {
// 			log.Println(i, "/", len(crns), semester.Code, crn)
// 			_, err := getCourseInfo(semester.Code, crn)
// 			if err != nil {
// 				log.Println(err)
// 			}
// 		}
// 	}
// }
