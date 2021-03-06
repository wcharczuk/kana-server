package model

import "github.com/blend/go-sdk/db/migration"

// Migrations returns the migration suite to bootstrap the database.
func Migrations() *migration.Suite {
	return migration.New(
		migration.OptGroups(
			quizzes(),
			quizResults(),
		),
	)
}

func users() *migration.Group {
	return migration.NewGroupWithAction(
		migration.TableNotExists("users"),
		migration.Statements(
			`CREATE TABLE users (
				id uuid not null primary key,
				created_utc timestampz not null,
				last_seen_utc timestampz not null,
				profile_id text not null,
				given_name text not null,
				family_name text not null,
				picture_url text not null,
				locale text not null,
				email text not null,
			)`,
		),
	)
}

func quizzes() *migration.Group {
	return migration.NewGroupWithAction(
		migration.TableNotExists("quiz"),
		migration.Statements(
			`CREATE TABLE quiz (
				id uuid not null primary key,
				created_utc timestampz not null,
				user_id uuid not null,
				hiragana boolean not null default false,	
				katakana boolean not null default false,	
				katakana_words boolean not null default false,	
				max_prompts int,	
				max_questions int,	
				max_repeat_history int,	
				prompts jsonb not null,
				prompt_weights jsonb,
				prompt_history jsonb
			)`,
			`ALTER TABLE quiz ADD CONSTRAINT fk_quiz_user_id FOREIGN KEY (user_id) REFERENCES users(id)`,
		),
	)
}

func quizResults() *migration.Group {
	return migration.NewGroupWithAction(
		migration.TableNotExists("quiz_result"),
		migration.Statements(
			`CREATE TABLE quiz_result (
				id uuid not null primary key,
				user_id uuid not null,
				quiz_id uuid not null,
				created_utc timestampz not null,
				answered_utc timestampz not null,
				prompt text not null,
				expected text not null,
				actual text not null
			)`,
			`ALTER TABLE quiz_result ADD CONSTRAINT fk_quiz_result_user_id FOREIGN KEY (user_id) REFERENCES users(id)`,
			`ALTER TABLE quiz_result ADD CONSTRAINT fk_quiz_result_quiz_id FOREIGN KEY (quiz_id) REFERENCES quiz(id)`,
		),
	)
}
