package main

import (
	"strings"
	"time"
)

type Program struct {
	Slug, Name, School, Degree, Duration, Summary, Description, Color, Icon string
	Modules, Careers                                                        []string
}

var Programs = []Program{
	{"computer-science", "Computer Science", "Computing & Technology", "B.Sc.", "4 years", "Build the technology that moves the world forward.", "Explore how computers work, learn to design reliable software, and turn ambitious ideas into useful digital products. Practical projects connect your learning to real problems from your first year.", "mint", "code", []string{"Programming fundamentals", "Algorithms & data structures", "Databases & cloud systems", "Artificial intelligence", "Software engineering", "Final-year project"}, []string{"Software engineer", "Data scientist", "Cybersecurity analyst"}},
	{"information-technology", "Information Technology", "Computing & Technology", "B.Sc.", "4 years", "Connect people, organisations, and possibilities.", "Learn to manage the systems that organisations depend on. Combine networking, security, databases, and practical problem solving in a programme built around applied technology.", "mint", "code", []string{"Computer networks", "Web development", "Database administration", "Information security", "Systems analysis", "IT project management"}, []string{"Systems administrator", "IT consultant", "Network engineer"}},
	{"business-administration", "Business Administration", "Business & Management", "B.Sc.", "4 years", "Learn to lead with purpose and make ideas happen.", "Understand how organisations grow, how teams work, and how good decisions are made. Develop your entrepreneurial thinking through case studies and collaborative projects.", "peach", "briefcase", []string{"Principles of management", "Marketing", "Business economics", "Financial management", "Entrepreneurship", "Strategic management"}, []string{"Business analyst", "Entrepreneur", "Operations manager"}},
	{"accounting", "Accounting", "Business & Management", "B.Sc.", "4 years", "Bring clarity, confidence, and integrity to business.", "Build a strong foundation in financial reporting, audit, taxation, and business ethics. Apply analytical thinking to financial decisions across different kinds of organisations.", "peach", "briefcase", []string{"Financial accounting", "Management accounting", "Audit & assurance", "Taxation", "Business law", "Financial analysis"}, []string{"Accountant", "Auditor", "Financial analyst"}},
	{"civil-engineering", "Civil Engineering", "Engineering & Design", "B.Eng.", "5 years", "Shape the places where tomorrow will happen.", "Study the design and construction of safe, sustainable infrastructure. Work with materials, structures, water systems, and transport through studio and laboratory projects.", "sand", "layers", []string{"Engineering mathematics", "Structural analysis", "Construction materials", "Geotechnical engineering", "Water resources", "Engineering design project"}, []string{"Civil engineer", "Structural designer", "Project engineer"}},
	{"electrical-engineering", "Electrical Engineering", "Engineering & Design", "B.Eng.", "5 years", "Power a more connected and sustainable future.", "Explore electrical systems, electronics, renewable energy, and automation. Combine scientific theory with hands-on design to solve challenges in energy and communication.", "sand", "layers", []string{"Circuit theory", "Digital electronics", "Power systems", "Control engineering", "Renewable energy", "Embedded systems"}, []string{"Electrical engineer", "Automation engineer", "Energy specialist"}},
	{"public-health", "Public Health", "Health & Life Sciences", "B.Sc.", "4 years", "Make healthier communities your life's work.", "Discover how research, education, and policy can improve community health. Learn to understand population needs and design practical responses to health challenges.", "lavender", "heart", []string{"Human biology", "Epidemiology", "Health promotion", "Biostatistics", "Environmental health", "Community research"}, []string{"Public health officer", "Health educator", "Research assistant"}},
	{"biological-sciences", "Biological Sciences", "Health & Life Sciences", "B.Sc.", "4 years", "Discover the living world, from cells to ecosystems.", "Investigate the science of life through laboratory work, field study, and research. Develop skills in observation, experimentation, and communicating scientific ideas.", "lavender", "heart", []string{"Cell biology", "Genetics", "Microbiology", "Ecology", "Biochemistry", "Research methods"}, []string{"Laboratory scientist", "Environmental researcher", "Biotechnology specialist"}},
	{"law", "Law", "Law & Social Sciences", "LL.B.", "5 years", "Ask better questions. Stand for a fairer world.", "Study legal principles, institutions, and their relationship with society. Develop careful reasoning, clear writing, and advocacy skills through guided case analysis and moot court.", "blue", "globe", []string{"Legal methods", "Constitutional law", "Contract law", "Criminal law", "Human rights", "Legal research & advocacy"}, []string{"Legal researcher", "Policy analyst", "Compliance officer"}},
	{"international-relations", "International Relations", "Law & Social Sciences", "B.A.", "4 years", "Understand our world and help bring it together.", "Examine diplomacy, global institutions, political economy, and international cooperation. Learn to assess different perspectives and communicate across cultures.", "blue", "globe", []string{"Political theory", "Diplomacy", "International organisations", "Global economics", "Conflict studies", "Foreign policy analysis"}, []string{"Policy researcher", "Diplomatic services", "Development coordinator"}},
	{"mass-communication", "Mass Communication", "Arts & Humanities", "B.A.", "4 years", "Tell the stories that deserve to be heard.", "Explore journalism, digital media, broadcasting, and public communication. Build a portfolio of thoughtful, engaging work while learning to report accurately and ethically.", "rose", "book", []string{"Media writing", "Broadcast production", "Digital storytelling", "Media law & ethics", "Public relations", "Multimedia project"}, []string{"Journalist", "Content producer", "Communications specialist"}},
	{"english-studies", "English Studies", "Arts & Humanities", "B.A.", "4 years", "Find your voice through language and literature.", "Explore texts, language, culture, and the power of expression. Strengthen your writing and critical thinking through close reading, discussion, and creative practice.", "rose", "book", []string{"Literary studies", "Linguistics", "Creative writing", "African literature", "Language & society", "Research & criticism"}, []string{"Editor", "Writer", "Language educator"}},
}

func FindProgram(slug string) (Program, bool) {
	for _, p := range Programs {
		if p.Slug == slug || p.Name == slug {
			return p, true
		}
	}
	return Program{}, false
}

func Schools() []string {
	var schools []string
	for _, p := range Programs {
		found := false
		for _, s := range schools {
			if s == p.School {
				found = true
				break
			}
		}
		if !found {
			schools = append(schools, p.School)
		}
	}
	return schools
}

type ScheduleCell struct{ Course, Room, Kind string }
type ScheduleRow struct {
	Time  string
	Cells []ScheduleCell
	Break bool
}

func ProgramSchedule(p Program, year int) []ScheduleRow {
	modules := append([]string{}, p.Modules...)
	if year == 1 {
		modules = []string{"Introduction to " + p.Name, "Academic writing", "Quantitative methods", p.Modules[0], "Study & research skills", "Digital literacy"}
	}
	rows := []ScheduleRow{}
	for i, hour := range []string{"08:00 – 10:00", "10:00 – 12:00", "12:00 – 13:00", "13:00 – 15:00", "15:00 – 17:00"} {
		row := ScheduleRow{Time: hour, Break: i == 2}
		for d := 0; d < 5; d++ {
			if i == 2 {
				row.Cells = append(row.Cells, ScheduleCell{Course: "Lunch & study break"})
				continue
			}
			if (i+d)%5 == 4 {
				row.Cells = append(row.Cells, ScheduleCell{Course: "Independent study", Kind: "Study"})
				continue
			}
			row.Cells = append(row.Cells, ScheduleCell{Course: modules[(i+d+year-1)%len(modules)], Room: []string{"Lecture Hall A", "Seminar Room 2", "Learning Lab", "Lecture Hall B", "Seminar Room 4"}[d], Kind: []string{"Lecture", "Seminar", "Practical", "Lecture", "Workshop"}[d]})
		}
		rows = append(rows, row)
	}
	return rows
}

func AcademicYear() string {
	now := time.Now()
	y := now.Year()
	if now.Month() < time.August {
		y--
	}
	return time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006") + "/" + time.Date(y+1, 1, 1, 0, 0, 0, 0, time.UTC).Format("06")
}
func contains(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}
