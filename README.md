# University of Excellence

A complete university website and staff workspace built with Go, server-rendered HTML, CSS, JavaScript, and SQLite. The original project has been expanded with a consistent responsive design and connected admissions and administration features.

## Run locally

You need Go 1.22 or newer and a C compiler because the SQLite driver uses CGO. Run these commands from the project directory:


Open **http://127.0.0.1:8080**. The staff portal is **http://127.0.0.1:8080/admin**, with username `school_website` and the password you set above. If no password is configured, the server creates a temporary password and prints it in the terminal. Sessions last eight hours and are cleared when the server restarts.

No Node.js installation or frontend build is required.

## What you can do

- Browse twelve undergraduate programmes across six academic schools, search by interest, and view programme subjects and career pathways.
- Read about the university, campus life, and faculty; filter faculty by school or name.
- Apply for a programme and receive a unique reference. Check an application's status with its reference and application email.
- Send an enquiry to the university. Submitted messages appear in the staff inbox.
- Read campus news and view upcoming events, with downloadable calendar invitations.
- View sample weekly timetables by programme and year, then print them.
- Search programmes, faculty, university pages, news, and upcoming events from one place.
- Use the staff dashboard to review applications, update admission statuses, export applications as CSV, manage messages, and add, edit, or delete faculty profiles, news stories, and events.






## Project structure

| File or directory | Responsibility |
| --- | --- |
| `main.go` | Routes, response headers, configuration, server startup |
| `catalog.go` | Programmes, academic schools, sample teaching schedules |
| `handlers.go` | Public pages, form validation, search, application tracking |
| `admin.go` | Sessions, staff access checks, form protection |
| `admin_login_handler.go` | Staff sign-in, dashboard, content editing, CSV exports |
| `helper.go` | SQLite queries, migrations, sample content |
| `html-file/` | Page templates and reusable components in `partials.html` |
| `static/` | Styles, progressive JavaScript, icons, photographs |
| `integration_test.go` | Application, administration, content, and migration tests |

