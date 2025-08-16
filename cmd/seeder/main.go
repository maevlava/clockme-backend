package main

import (
	"context"
	"database/sql"
	"flag"
	"github.com/clockme/clockme-backend/internal/features/auth"
	db2 "github.com/clockme/clockme-backend/internal/shared/db"
	"github.com/clockme/clockme-backend/internal/shared/logger"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"math/rand"
	"os"
	"time"
)

func main() {
	logger.Init()
	ctxBg := context.Background()

	seedCmd := flag.String("seed", "all", "all, users, projects, tasks, records, or links")
	flag.Parse()

	conn, queries := connectDB()
	defer conn.Close()

	switch *seedCmd {
	case "all":
		seedUsers(queries, ctxBg)
		seedProjects(queries, ctxBg)
		seedProjectsUsers(queries, ctxBg, seedUsers(queries, ctxBg), seedProjects(queries, ctxBg))
		seedTasks(queries, ctxBg)
		seedTimeRecords(queries, ctxBg)
		break
	case "users":
		seedUsers(queries, ctxBg)
		break
	case "projects":
		seedProjects(queries, ctxBg)
	case "tasks":
		seedTasks(queries, ctxBg)
		break
	case "records":
		seedTimeRecords(queries, ctxBg)
		break
	case "links":
		log.Info().Msg("Seeding links only. Fetching existing data...")
		userIDs, err := getAllUserIDs(queries, ctxBg)
		if err != nil {
			log.Fatal().Err(err).Msg("Could not fetch user IDs")
		}
		projectIDs, err := getAllProjectIDs(queries, ctxBg)
		if err != nil {
			log.Fatal().Err(err).Msg("Could not fetch project IDs")
		}
		if len(userIDs) > 0 && len(projectIDs) > 0 {
			seedProjectsUsers(queries, ctxBg, userIDs, projectIDs)
		} else {
			log.Warn().Msg("No users or projects found in the database to link.")
		}
	default:
		log.Error().Msgf("Unknown seed command: %s", *seedCmd)
	}

}

type TaskInfo struct {
	TaskID    uuid.UUID
	ProjectID uuid.UUID
}

func connectDB() (*sql.DB, *db2.Queries) {
	ctxBg := context.Background()

	dbSource := os.Getenv("DB_SOURCE")
	if dbSource == "" {
		log.Fatal().Msg("DB_SOURCE is not set")
	}

	conn, err := sql.Open("postgres", dbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	ctx, cancel := context.WithTimeout(ctxBg, 10*time.Second)
	defer cancel()
	if err := conn.PingContext(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping database")
	}

	log.Info().Msg("Successfully connected to the database!")

	queries := db2.New(conn)

	return conn, queries
}
func seedUsers(queries *db2.Queries, ctxBg context.Context) []uuid.UUID {
	var userIDs []uuid.UUID

	for i := 0; i < 3; i++ {
		hashedPassword, err := auth.HashPassword(faker.Password())
		if err != nil {
			log.Error().Err(err).Msg("Failed to create user")
		}
		newUser, err := queries.CreateUser(ctxBg, db2.CreateUserParams{
			ID:             uuid.New(),
			Name:           faker.Name(),
			HashedPassword: hashedPassword,
			Email:          faker.Email(),
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to create user")
		} else {
			log.Info().
				Stringer("id", newUser.ID).
				Str("name", newUser.Name).
				Str("email", newUser.Email).
				Msg("Created user")
			userIDs = append(userIDs, newUser.ID)
		}
	}
	return userIDs
}
func seedProjects(queries *db2.Queries, ctxBg context.Context) []uuid.UUID {
	var projectIDs []uuid.UUID

	for i := 0; i < 3; i++ {
		newProject, err := queries.CreateProject(ctxBg, db2.CreateProjectParams{
			ID:   uuid.New(),
			Name: faker.Name(),
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to create project")
		} else {
			log.Info().
				Stringer("id", newProject.ID).
				Str("name", newProject.Name).
				Msg("Created project")
			projectIDs = append(projectIDs, newProject.ID)
		}
	}
	return projectIDs
}
func seedProjectsUsers(queries *db2.Queries, ctxBg context.Context, userIDs []uuid.UUID, projectIDs []uuid.UUID) {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for _, projectID := range projectIDs {
		// how many users will be added
		numUserToAdd := r.Intn(len(userIDs)) + 1

		//shuffle the users
		r.Shuffle(len(userIDs), func(i, j int) {
			userIDs[i], userIDs[j] = userIDs[j], userIDs[i]
		})

		for i := 0; i < numUserToAdd; i++ {
			userID := userIDs[i]
			_, err := queries.AddUserToProject(ctxBg, db2.AddUserToProjectParams{
				UserID:    userID,
				ProjectID: projectID,
			})
			if err != nil {
				log.Warn().Err(err).
					Stringer("user id", userID).
					Stringer("project id", projectID).
					Msg("Failed to add user to project might be duplicate")

				continue
			}
			log.Info().
				Stringer("user id", userID).
				Stringer("project id", projectID).
				Msg("Added user to project")
		}
	}
}
func seedTasks(queries *db2.Queries, ctxBg context.Context) []uuid.UUID {
	var taskIDs []uuid.UUID

	projectIDs, err := getAllProjectIDs(queries, ctxBg)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not fetch project IDs")
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < 8; i++ {
		randomIndex := r.Intn(len(projectIDs))

		randomProjectID := projectIDs[randomIndex]

		newTask, err := queries.CreateTask(ctxBg, db2.CreateTaskParams{
			ID:        uuid.New(),
			Name:      faker.Sentence(),
			ProjectID: randomProjectID,
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to create task")
		} else {
			log.Info().
				Stringer("id", newTask.ID).
				Str("name", newTask.Name).
				Msg("Created task")
			taskIDs = append(taskIDs, newTask.ID)
		}
	}

	return taskIDs
}
func seedTimeRecords(queries *db2.Queries, ctxBg context.Context) []uuid.UUID {
	var timeRecordsIDs []uuid.UUID

	tasksInfo, err := getAllTasksInfo(queries, ctxBg)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not fetch tasks info")
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	now := time.Now()
	tenDaysAgo := now.AddDate(0, 0, -10)
	timeWindowInSeconds := now.Unix() - tenDaysAgo.Unix()
	for i := 0; i < 15; i++ {
		randomTaskInfoIndex := r.Intn(len(tasksInfo))
		randomTaskInfo := tasksInfo[randomTaskInfoIndex]

		randomOffsetInSeconds := r.Int63n(timeWindowInSeconds)
		startTime := tenDaysAgo.Add(time.Second * time.Duration(randomOffsetInSeconds))

		minDurationInSeconds := 5 * 60
		maxDurationInSeconds := 6 * 60 * 60
		randomDurationInSeconds := r.Intn(maxDurationInSeconds-minDurationInSeconds) + minDurationInSeconds
		endTime := startTime.Add(time.Second * time.Duration(randomDurationInSeconds))
		params := db2.CreateTimeRecordParams{
			ID:        uuid.New(),
			Name:      faker.Sentence(),
			StartTime: startTime,
			EndTime:   endTime,
			TaskID:    randomTaskInfo.TaskID,
			ProjectID: randomTaskInfo.ProjectID,
		}
		newRecord, err := queries.CreateTimeRecord(ctxBg, params)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create time record")
			continue
		} else {
			log.Info().
				Stringer("id", newRecord.ID).
				Str("name", newRecord.Name).
				Msg("Created time record")
			timeRecordsIDs = append(timeRecordsIDs, newRecord.ID)
		}
	}
	return timeRecordsIDs

}
func getAllUserIDs(queries *db2.Queries, ctxBg context.Context) ([]uuid.UUID, error) {
	users, err := queries.GetAllUsers(ctxBg)
	if err != nil {
		return nil, err
	}

	userIDs := make([]uuid.UUID, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}
	return userIDs, nil
}
func getAllProjectIDs(queries *db2.Queries, ctxBg context.Context) ([]uuid.UUID, error) {
	projects, err := queries.GetAllProjects(ctxBg)
	if err != nil {
		return nil, err
	}

	projectIDs := make([]uuid.UUID, len(projects))
	for i, p := range projects {
		projectIDs[i] = p.ID
	}
	return projectIDs, nil
}
func getAllTasksInfo(queries *db2.Queries, ctxBg context.Context) ([]TaskInfo, error) {
	tasks, err := queries.GetAllTasks(ctxBg)
	if err != nil {
		return nil, err
	}
	taskInfo := make([]TaskInfo, len(tasks))
	for i, t := range tasks {
		taskInfo[i] = TaskInfo{
			TaskID:    t.ID,
			ProjectID: t.ProjectID,
		}
	}
	return taskInfo, nil
}
