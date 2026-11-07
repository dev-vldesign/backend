package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Questionnaire struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID primitive.ObjectID `bson:"project_id" json:"project_id"`
	Questions []Question         `bson:"questions" json:"questions"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type Question struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Text             string             `bson:"text" json:"text"`
	Answer           []string           `bson:"answer,omitempty" json:"answer,omitempty"`
	OtherAnswer      string             `bson:"other_answer,omitempty" json:"other_answer,omitempty"`
	Documents        []Document         `bson:"documents,omitempty" json:"documents,omitempty"`
	PossibleAnswers  []string           `bson:"possible_answers,omitempty" json:"possible_answers,omitempty"`
	Type             string             `bson:"type" json:"type"`
	IsAnswerRejected bool               `bson:"is_answer_rejected" json:"is_answer_rejected"`
	HelpText         []string           `bson:"help_text" json:"help_text"`
	HelperVideo      string             `bson:"helper_video" json:"helper_video"`
}
