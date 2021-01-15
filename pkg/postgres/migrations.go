package postgres

import "github.com/blend/go-sdk/db/migration"

func Migrations() *migration.Suite {
	return migration.New(
		migration.OptGroups(
			quizzes(),
			quizResults(),
		),
	)
}

func quizzes() *migration.Group {
	return migration.NewGroupWithAction(
		migration.TableNotExists("quiz"),
		migration.Statements(
			`CREATE TABLE quiz (
				id uuid not null primary key,
				created_utc timestamp not null,
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
		),
	)
}

func quizResults() *migration.Group {
	return migration.NewGroupWithAction(
		migration.TableNotExists("quiz_result"),
		migration.Statements(
			`CREATE TABLE quiz_result (
				id uuid not null primary key,
				quiz_id uuid not null,
				created_utc timestamp not null,
				answered_utc timestamp not null,
				prompt text not null,
				expected text not null,
				actual text not null
			)`,
			`ALTER TABLE quiz_result ADD CONSTRAINT fk_quiz_result_quiz_id FOREIGN KEY (quiz_id) REFERENCES quiz(id)`,
		),
	)
}
