package repository

import (
	"context"
	"egogoger/internal/pkg/models"
	"egogoger/internal/pkg/service"
	"github.com/jackc/pgx/v4/pgxpool"

	//"log"
	"net/http"
)

type serviceRepository struct {
	db *pgxpool.Pool
}

func NewPgxServiceRepository(db *pgxpool.Pool) service.Repository {
	return &serviceRepository{db: db}
}

func (sr *serviceRepository) TruncateAll() int {
	sqlStatement := `
		TRUNCATE usr CASCADE;`
	_, err := sr.db.Exec(context.Background(), sqlStatement)
	if err != nil {
		//log.Println("ERROR: Service Repo TruncateAll")
		return http.StatusInternalServerError
	}

	sqlStatement = `
		UPDATE summary SET users = 0, forums = 0, threads = 0, posts = 0 WHERE users != -1;`
	_, err = sr.db.Exec(context.Background(), sqlStatement)
	if err != nil {
		//log.Println("ERROR: Service Repo TruncateAll")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func (sr *serviceRepository) GetInfo() (*models.Summary, int) {
	sqlStatement := `
		UPDATE summary
		SET users = GREATEST(uid.coalesce, users),
			forums = GREATEST(fid.coalesce, forums),
			threads = GREATEST(tid.coalesce, threads),
			posts = GREATEST(pid.coalesce, posts)
			FROM
				(SELECT COALESCE(COUNT(*), 0) FROM usr)    AS uid,
				(SELECT COALESCE(COUNT(*), 0) FROM forum)  AS fid,
				(SELECT COALESCE(MAX(id), 0) FROM thread) AS tid,
				(SELECT COALESCE(MAX(id), 0) FROM post)   AS pid
		RETURNING users, forums, threads, posts;`

	summary := new(models.Summary)
	err := sr.db.QueryRow(context.Background(), sqlStatement).Scan(
		&summary.Users,
		&summary.Forums,
		&summary.Threads,
		&summary.Posts)
	if err != nil {
		//log.Println("ERROR: Service Repo GetInfo")
		return nil, http.StatusInternalServerError
	} else {
		return summary, http.StatusOK
	}
}
