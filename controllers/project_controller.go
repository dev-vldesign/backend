package controllers

import (
	"fmt"
	"math"
	"net/http"

	"context"
	"time"

	"github.com/Msaad668/vl-design-backend/db"
	"github.com/Msaad668/vl-design-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CreateProjectRequest struct {
	Address               string   `json:"address" binding:"required"`
	CustomerNeedsHelpWith string   `json:"customer_needs_help_with" binding:"required"`
	ProjectArea           string   `json:"project_area" binding:"required"`
	PropertySize          string   `json:"property_size" binding:"required"`
	UniqueChallenges      []string `json:"unique_challenges" binding:"required"`
	YardFeatures          []string `json:"yard_features" binding:"required"`
	ProjectTimeline       string   `json:"project_timeline" binding:"required"`
	SelectedPackage       string   `json:"selected_package" binding:"required"`
}

func CreateProject(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	customerID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate := time.Now()
	endDate := startDate.AddDate(0, 1, 0) // +1 month

	defaultTasks := createDefaultProjectTasks()

	newProject := models.Project{
		ID:                    primitive.NewObjectID(),
		Address:               req.Address,
		CustomerNeedsHelpWith: req.CustomerNeedsHelpWith,
		ProjectArea:           req.ProjectArea,
		PropertySize:          req.PropertySize,
		UniqueChallenges:      req.UniqueChallenges,
		YardFeatures:          req.YardFeatures,
		ProjectTimeline:       req.ProjectTimeline,
		SelectedPackage:       req.SelectedPackage,

		Status:     "Pending", // Default
		IsPaid:     false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		CustomerID: customerID,

		DesignerID:       nil,
		SecondDesignerID: nil,
		OrderID:          nil,

		ProjectDuration:  "1 month",
		ProjectStartDate: startDate,
		ProjectEndDate:   endDate,

		ProjectTasks: defaultTasks,
	}

	collection := db.GetCollection(db.Client, "projects")
	_, err = collection.InsertOne(c, newProject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	err = createDefaultQuestionnaireForProject(c, newProject.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create questionnaire"})
		return
	}

	c.JSON(http.StatusCreated, newProject)
}

func createDefaultProjectTasks() []models.ProjectTask {
	tasks := []string{
		"Quesionnaire finalized",
		"Inhouse 1st draft",
		"Final draft",
		"Client’s Feedback",
		"Final Package",
	}

	draftChildTasks := []string{
		"2D Layout",
		"3D render",
		"REV(A)",
	}

	var projectTasks []models.ProjectTask

	currentStart := time.Now()
	oneWeekLater := func(t time.Time) time.Time {
		return t.AddDate(0, 0, 7)
	}

	for i, tName := range tasks {
		end := oneWeekLater(currentStart)
		parentTask := models.ProjectTask{
			ID:            primitive.NewObjectID(),
			TaskName:      tName,
			TaskDuration:  "1 week",
			TaskStartDate: currentStart,
			TaskEndDate:   end,
			TaskStatus:    "Not Started",
			Children:      nil,
		}

		currentStart = end

		if i == 1 || i == 2 {
			var children []models.ProjectTask
			childCurrentStart := parentTask.TaskEndDate

			for _, childName := range draftChildTasks {
				childEnd := oneWeekLater(childCurrentStart)
				childTask := models.ProjectTask{
					ID:            primitive.NewObjectID(),
					TaskName:      childName,
					TaskDuration:  "1 week",
					TaskStartDate: childCurrentStart,
					TaskEndDate:   childEnd,
					TaskStatus:    "Not Started",
				}
				children = append(children, childTask)
				childCurrentStart = childEnd
			}
			parentTask.Children = children
		}

		projectTasks = append(projectTasks, parentTask)
	}

	return projectTasks
}

func createDefaultQuestionnaireForProject(c *gin.Context, projectID primitive.ObjectID) error {
	questions := []models.Question{
		{
			ID:              primitive.NewObjectID(),
			Text:            "1- Why do you want to design your yard? There could be more than one reason, so please select all that apply to you. If you have additional reasons, be sure to note those as well.",
			PossibleAnswers: []string{"Just moved to a new house", "Baby on the way", "Welcoming our new furry friends", "Haven't updated my yard in years.", "Kids have grown up. Want a fresh look.", "Other"},
			Type:            "CheckBox",
		},
		{
			ID:              primitive.NewObjectID(),
			Text:            "2- What's on your design wish list? These are the elements you wish to incorporate into your landscape design. We'll strive to include all your selections. Please provide images of each design element in your inspiration photo collection, along with a brief explanation for clarity. Please add anything else that you want in your landscape.?",
			PossibleAnswers: []string{"Plants", "Pool", "Spa", "Play Area", "Outdoor Kitchen", "Outdoor Furniture", "Deck", "Exterior paint", "Pergola", "Dog Aera", "Storage Aera", "Drainage", "Retaining Walls", "Lighting", "Fencing", "Firepit", "Shed", "Other"},
			Type:            "CheckBox",
		},
		{
			ID:   primitive.NewObjectID(),
			Text: "3- Priority List!! List the elements that are absolute must-haves for your landscape design—no negotiation. We'll ensure these are prioritised first, and then we'll aim to incorporate the rest of your wishlist items if space permits.",
			Type: "TextInput",
		},
		{
			ID:              primitive.NewObjectID(),
			Text:            "4- Each yard is unique! What are the unique challenges in your yard?",
			PossibleAnswers: []string{"Sloped", "Poor Soil", "Lack of Privacy", "Full Shade", "Clear Slate", "Difficult terrain", "Disjointed style", "Overgrown", "Unhappy Plants", "Wasted Space", "Other"},
		},
		{
			ID:   primitive.NewObjectID(),
			Text: "5- In order to create a landscape design that perfectly aligns with your needs, as well as those of your family and friends, we kindly ask for your input. Please take a moment to select the options that best reflect your preferences. Additional comments or details are welcome as they will help us ensure the utmost precision in customisation.",
			Type: "TextInput",
		},
		{
			ID:   primitive.NewObjectID(),
			Text: "6- Lifestyle.",
			PossibleAnswers: []string{
				"Love entertaining need a large space for my family and friends to enjoy",
				"Like quiet and relaxing time outdoors.",
				"Very outdoorsy need plenty of room to hang out.",
				"I have pets that need plenty of space outside",
				"I want to grow some of my own produce.",
				"Time to Spend in the garden",
				"Garden style",
				"Other",
			},
		},
		{
			ID:   primitive.NewObjectID(),
			Text: "7- Time to spend in the garden?",
			PossibleAnswers: []string{
				"Have plenty of time to maintain gardens or planning on getting someone to help with maintenance.",
				"Have a bit of time to garden so can do some maintenance.",
				"No time for maintenance create a garden as maintenance free as possible.",
			},
		},
		{
			ID:   primitive.NewObjectID(),
			Text: "8- Please list any specific plants you love and would want included",
			Type: "TextInput",
		},
		{
			ID:   primitive.NewObjectID(),
			Text: "9- Please list any plants you dislike and would not want included",
			Type: "TextInput",
		},
		{
			ID:   primitive.NewObjectID(),
			Text: "10- What is the site soil condition? If you have existing gardens how are they going? Do you have trouble growing plants in the past due to the soil condition. Describe in a few words what is the problem with you soil ie rocky, clay, black soil, boggy?",
			Type: "TextInput",
		},
		{
			ID:   primitive.NewObjectID(),
			Text: "11- Please select how you would like your design laid out. Straight and symmetrical or curved and sweeping.?",
			PossibleAnswers: []string{
				"Curves.",
				"Straight line.",
				"Both.",
			},
		},
		{
			ID:   primitive.NewObjectID(),
			Text: "12- Garden Style, Have a look at the examples below and let us know what style of garden you like. You can choose more than one.?",
			PossibleAnswers: []string{
				"Formal.",
				"Cottage.",
				"Rambling Parks and Gardens.",
				"Australian Native.",
				"Waterwise.",
				"Mediterranean.",
				"Architectural.",
			},
		},
		{
			ID:   primitive.NewObjectID(),
			Text: "13- Please share anything else that you would like to add to help us design the perfect landscape for you?",
			Type: "TextInput",
		},
		{
			ID:   primitive.NewObjectID(),
			Text: "14- Pool size and shape",
			Type: "TextInput",
		},
		{
			ID:       primitive.NewObjectID(),
			Text:     "14- Already chosen a pool/spa?",
			HelpText: []string{"If you would like a pool/spa included in your design please select from below the options."},
			Type:     "TextInput",
		},
		{
			ID:   primitive.NewObjectID(),
			Text: "15- Upload images here",
			HelpText: []string{
				"PLEASE WATCH ABOVE VIDEO PRIOR TO TAKING THE PHOTOS OF YOUR PROPERTY!!.",
				"Please upload extensive site photos of the area that is to be designed. Please ensure the photos are clear and are taken at eye level and straight. These are used to create accurate render images of your home with your new landscape design featured. The clearer they are the more accurate the images will be!!.",
			},
			Type: "RichQuestion",
		},
		{
			ID:   primitive.NewObjectID(),
			Text: "16- Upload videos here",
			HelpText: []string{
				"PLEASE WATCH ABOVE VIDEO PRIOR TO VIDEOING YOUR PROPERTY!!",
				"Please provide a walkthrough video with commentary. This video should explain the site we will be designing for you. It is important that we can clearly view the area while you discuss any problems you're encountering and point out any existing features you want to retain or remove. It's preferable to upload several videos, each no longer than five minutes, rather than one lengthy video. Not required on under construction buildings. #Essential for existing buildings.",
			},
			HelperVideo: "https://www.youtube.com/embed/KuLKdJHXopc",
			Type:        "RichQuestion",
		},
		{
			ID:   primitive.NewObjectID(),
			Text: "17- Mudmap",
			Type: "RichQuestion",
		},
	}

	questionnaire := models.Questionnaire{
		ID:        primitive.NewObjectID(),
		ProjectID: projectID,
		Questions: questions,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	questionnaireCollection := db.GetCollection(db.Client, "questionnaires")
	_, err := questionnaireCollection.InsertOne(c, questionnaire)
	return err
}

func GetProjectsForUser(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	collection := db.GetCollection(db.Client, "projects")
	var filter bson.M

	switch role {
	case "admin":
		filter = bson.M{}
	case "designer":
		filter = bson.M{"designer_id": userObjectID}
	case "customer":
		filter = bson.M{"customer_id": userObjectID}
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{"$match": filter},
		// Convert IDs to ObjectID format for lookups
		{
			"$addFields": bson.M{
				"customer_id_object": bson.M{
					"$convert": bson.M{
						"input":   "$customer_id",
						"to":      "objectId",
						"onError": nil,
						"onNull":  nil,
					},
				},
				"designer_id_object": bson.M{
					"$convert": bson.M{
						"input":   "$designer_id",
						"to":      "objectId",
						"onError": nil,
						"onNull":  nil,
					},
				},
				"second_designer_id_object": bson.M{
					"$convert": bson.M{
						"input":   "$second_designer_id",
						"to":      "objectId",
						"onError": nil,
						"onNull":  nil,
					},
				},
			},
		},
		// Lookup customer info (only first_name and last_name)
		{
			"$lookup": bson.M{
				"from": "users",
				"let":  bson.M{"customerId": "$customer_id_object"},
				"pipeline": []bson.M{
					{"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$_id", "$$customerId"}}}},
					{"$project": bson.M{
						"first_name": 1,
						"last_name":  1,
						"_id":        0,
					}},
				},
				"as": "customer_info",
			},
		},
		// Lookup designer info (only first_name and last_name)
		{
			"$lookup": bson.M{
				"from": "users",
				"let":  bson.M{"designerId": "$designer_id_object"},
				"pipeline": []bson.M{
					{"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$_id", "$$designerId"}}}},
					{"$project": bson.M{
						"first_name": 1,
						"last_name":  1,
						"_id":        0,
					}},
				},
				"as": "designer_info",
			},
		},
		// Lookup second designer info (only first_name and last_name)
		{
			"$lookup": bson.M{
				"from": "users",
				"let":  bson.M{"secondDesignerId": "$second_designer_id_object"},
				"pipeline": []bson.M{
					{"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$_id", "$$secondDesignerId"}}}},
					{"$project": bson.M{
						"first_name": 1,
						"last_name":  1,
						"_id":        0,
					}},
				},
				"as": "second_designer_info",
			},
		},
		// Unwind all lookups (preserve null for optional fields)
		{
			"$unwind": bson.M{
				"path":                       "$customer_info",
				"preserveNullAndEmptyArrays": true,
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$designer_info",
				"preserveNullAndEmptyArrays": true,
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$second_designer_info",
				"preserveNullAndEmptyArrays": true,
			},
		},
		// Create full names
		{
			"$addFields": bson.M{
				"customer_name": bson.M{
					"$cond": bson.M{
						"if": bson.M{"$and": []interface{}{
							bson.M{"$ifNull": []interface{}{"$customer_info.first_name", false}},
							bson.M{"$ifNull": []interface{}{"$customer_info.last_name", false}},
						}},
						"then": bson.M{"$concat": bson.A{
							"$customer_info.first_name",
							" ",
							"$customer_info.last_name",
						}},
						"else": "$customer_info.first_name",
					},
				},
				"designer_name": bson.M{
					"$cond": bson.M{
						"if": bson.M{"$and": []interface{}{
							bson.M{"$ifNull": []interface{}{"$designer_info.first_name", false}},
							bson.M{"$ifNull": []interface{}{"$designer_info.last_name", false}},
						}},
						"then": bson.M{"$concat": bson.A{
							"$designer_info.first_name",
							" ",
							"$designer_info.last_name",
						}},
						"else": "$designer_info.first_name",
					},
				},
				"second_designer_name": bson.M{
					"$cond": bson.M{
						"if": bson.M{"$and": []interface{}{
							bson.M{"$ifNull": []interface{}{"$second_designer_info.first_name", false}},
							bson.M{"$ifNull": []interface{}{"$second_designer_info.last_name", false}},
						}},
						"then": bson.M{"$concat": bson.A{
							"$second_designer_info.first_name",
							" ",
							"$second_designer_info.last_name",
						}},
						"else": "$second_designer_info.first_name",
					},
				},
			},
		},
		{
        "$sort": bson.M{
            "created_at": -1,  // -1 for descending order (newest first)
        	},
    	},
		// Final projection - only include the fields we want
		{
			"$project": bson.M{
				"_id":                  1,
				"project_start_date":   1,
				"project_end_date":     1,
				"status":               1,
				"customer_name":        1,
				"designer_name":        1,
				"second_designer_name": 1,
			},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "No projects found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve projects"})
		return
	}
	defer cursor.Close(ctx)

	var projects []bson.M
	if err := cursor.All(ctx, &projects); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse projects"})
		return
	}

	c.JSON(http.StatusOK, projects)
}

func GetProjectByID(c *gin.Context) {
	projectIDStr := c.Param("projectID")
	projectObjectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	projectCollection := db.GetCollection(db.Client, "projects")
	var project models.Project
	if err := projectCollection.FindOne(context.Background(), bson.M{"_id": projectObjectID}).Decode(&project); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project"})
		return
	}

	totalTasks, doneTasks, inProgressTasks, notStartedTasks := getTaskStats(project.ProjectTasks)
	progressPercent := 0.0
	if totalTasks > 0 {
		progressPercent = (float64(doneTasks) / float64(totalTasks)) * 100
	}

	elapsedDurationDays := math.Floor(time.Since(project.ProjectStartDate).Hours() / 24)
	if elapsedDurationDays < 0 {
		elapsedDurationDays = 0 // if start date is in the future, keep it 0
	}

	userIDsSet := make(map[primitive.ObjectID]struct{})

	if project.DesignerID != nil {
		userIDsSet[*project.DesignerID] = struct{}{}
	}
	if project.SecondDesignerID != nil {
		userIDsSet[*project.SecondDesignerID] = struct{}{}
	}

	collectTaskAssignees(project.ProjectTasks, userIDsSet)

	var userIDs []primitive.ObjectID
	for k := range userIDsSet {
		userIDs = append(userIDs, k)
	}

	assigneesInfo, err := fetchUsersByIDs(userIDs)
	if err != nil {
		fmt.Println("Could not fetch user info for assignees:", err)
	}

	var clientInfo *models.User
	if !project.CustomerID.IsZero() {
		singleUser, err := fetchUserByID(project.CustomerID)
		if err == nil {
			clientInfo = &singleUser
		} else {
			fmt.Println("Could not fetch client info:", err)
		}
	}

	response := gin.H{
		"project_id":  project.ID.Hex(),
		"address":     project.Address,
		"client_info": nil,

		"customer_needs_help_with": project.CustomerNeedsHelpWith,
		"project_area":             project.ProjectArea,
		"property_size":            project.PropertySize,
		"unique_challenges":        project.UniqueChallenges,
		"yard_features":            project.YardFeatures,
		"project_timeline":         project.ProjectTimeline,
		"selected_package":         project.SelectedPackage,
		"status":                   project.Status,
		"is_paid":                  project.IsPaid,
		"created_at":               project.CreatedAt,
		"updated_at":               project.UpdatedAt,

		"designer_id":        nil,
		"second_designer_id": nil,
		"order_id":           nil,

		"project_duration":   project.ProjectDuration,
		"project_start_date": project.ProjectStartDate,
		"project_end_date":   project.ProjectEndDate,

		"total_tasks":       totalTasks,
		"done_tasks":        doneTasks,
		"in_progress_tasks": inProgressTasks,
		"not_started_tasks": notStartedTasks,
		"progress_percent":  math.Round(progressPercent),

		"elapsed_duration_days": elapsedDurationDays,
		"assignees_info":        assigneesInfo,
		"project_tasks":         project.ProjectTasks,
	}

	if clientInfo != nil {
		response["client_info"] = gin.H{
			"id":         clientInfo.ID.Hex(),
			"first_name": clientInfo.FirstName,
			"last_name":  clientInfo.LastName,
			"email":      clientInfo.Email,
			"role":       clientInfo.Role,
			"phone":      clientInfo.Phone,
		}
	}

	if project.DesignerID != nil {
		response["designer_id"] = project.DesignerID.Hex()
	}
	if project.SecondDesignerID != nil {
		response["second_designer_id"] = project.SecondDesignerID.Hex()
	}
	if project.OrderID != nil {
		response["order_id"] = project.OrderID.Hex()
	}

	c.JSON(http.StatusOK, response)
}

func getTaskStats(tasks []models.ProjectTask) (total, done, inProgress, notStarted int) {
	for _, t := range tasks {
		total++
		switch t.TaskStatus {
		case "Completed", "Done":
			done++
		case "In Progress":
			inProgress++
		case "Not Started":
			notStarted++
		default:
			notStarted++
		}
		ct, cd, ci, cn := getTaskStats(t.Children)
		total += ct
		done += cd
		inProgress += ci
		notStarted += cn
	}
	return
}

func collectTaskAssignees(tasks []models.ProjectTask, userIDsSet map[primitive.ObjectID]struct{}) {
	for _, t := range tasks {
		if t.TaskAssignee != nil {
			userIDsSet[*t.TaskAssignee] = struct{}{}
		}
		if len(t.Children) > 0 {
			collectTaskAssignees(t.Children, userIDsSet)
		}
	}
}

func fetchUsersByIDs(ids []primitive.ObjectID) ([]models.User, error) {
	if len(ids) == 0 {
		return []models.User{}, nil
	}

	collection := db.GetCollection(db.Client, "users")
	cursor, err := collection.Find(context.Background(), bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var users []models.User
	if err := cursor.All(context.Background(), &users); err != nil {
		return nil, err
	}
	return users, nil
}

func fetchUserByID(userID primitive.ObjectID) (models.User, error) {
	var user models.User
	collection := db.GetCollection(db.Client, "users")
	err := collection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	return user, err
}


// DeleteProject deletes a project by its ID
func DeleteProject(c *gin.Context) {
    projectID := c.Param("projectID")
    
    // Convert string ID to ObjectID
    projectObjectID, err := primitive.ObjectIDFromHex(projectID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID format"})
        return
    }

    // Get the projects collection
    collection := db.GetCollection(db.Client, "projects")

    // Delete the project
    result, err := collection.DeleteOne(context.TODO(), bson.M{"_id": projectObjectID})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project", "details": err.Error()})
        return
    }

    // Check if a project was actually deleted
    if result.DeletedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
        return
    }

    // Optionally: Delete related data (questionnaires, tasks, etc.)
    // This depends on your application's requirements
    // Example:
    // questionnaireCollection := db.GetCollection(db.Client, "questionnaires")
    // _, _ = questionnaireCollection.DeleteMany(context.TODO(), bson.M{"project_id": projectObjectID})

    c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}