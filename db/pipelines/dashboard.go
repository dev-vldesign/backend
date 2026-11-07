package pipelines

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetDashboardData(userObjectID primitive.ObjectID) []bson.M {
	DashboardData := []bson.M{
		// Match projects for this designer
		{
			"$match": bson.M{
				"$or": []bson.M{
					{"designer_id": userObjectID},
					{"second_designer_id": userObjectID},
				},
				"project_start_date": bson.M{
					"$gte": time.Now().AddDate(0, 0, -360),
				},
			},
		},
		// Convert IDs for lookups
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
		// Create full names and project necessary fields
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
				"month": bson.M{"$month": "$project_start_date"},
				"year":  bson.M{"$year": "$project_start_date"},
				"is_active": bson.M{
					"$and": []bson.M{
						{"$ne": []interface{}{"$status", "completed"}},
						{"$ne": []interface{}{"$status", "archived"}},
					},
				},
				"is_overdue": bson.M{
					"$and": []bson.M{
						{"$ne": []interface{}{"$status", "completed"}},
						{"$lt": []interface{}{"$project_end_date", time.Now()}},
					},
				},
			},
		},
		// Group for all required metrics
		{
			"$facet": bson.M{
				"status_summary": []bson.M{
					{
						"$group": bson.M{
							"_id": nil,
							"total_projects": bson.M{
								"$sum": bson.M{
									"$cond": []interface{}{
										bson.M{"$or": []interface{}{
											bson.M{"$eq": []interface{}{"$status", "completed"}},
											bson.M{"$eq": []interface{}{"$status", "done"}},
										}},
										1,
										0,
									},
								},
							},
							"in_progress": bson.M{
								"$sum": bson.M{"$cond": []interface{}{"$is_active", 1, 0}},
							},
							"overdue": bson.M{
								"$sum": bson.M{"$cond": []interface{}{"$is_overdue", 1, 0}},
							},
						},
					},
					{
						"$addFields": bson.M{
							"workload": "$in_progress",
						},
					},
				},
				"performance_graph": []bson.M{
					// First get the current year
					{
						"$addFields": bson.M{
							"currentYear": bson.M{"$year": time.Now()},
						},
					},
					// Group by month and year to get actual counts
					{
						"$group": bson.M{
							"_id": bson.M{
								"month": "$month",
								"year":  "$year",
							},
							"count": bson.M{"$sum": 1},
						},
					},
					// Generate all months for the current year
					{
						"$unionWith": bson.M{
							"coll": "projects",
							"pipeline": []bson.M{
								{
									"$project": bson.M{
										"months": bson.M{"$range": []interface{}{1, 13}},
									},
								},
								{"$unwind": "$months"},
								{
									"$project": bson.M{
										"_id": bson.M{
											"month": "$months",
											"year":  bson.M{"$year": time.Now()},
										},
										"count": bson.M{"$literal": 0},
									},
								},
							},
						},
					},
					// Group again to combine actual and generated months
					{
						"$group": bson.M{
							"_id":   "$_id",
							"count": bson.M{"$max": "$count"},
						},
					},
					// Sort by year and month
					{"$sort": bson.M{"_id.year": 1, "_id.month": 1}},
					{"$limit": 12},
				},
				"recent_projects": []bson.M{
					{"$sort": bson.M{"project_start_date": -1}},
					{"$limit": 5},
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
				},
				"current_tasks": []bson.M{
					{"$unwind": "$project_tasks"},
					{
						"$match": bson.M{
							"project_tasks.task_status": bson.M{"$ne": "done"},
							"$or": []bson.M{
								{"project_tasks.task_assignee": userObjectID},
								{"project_tasks.task_assignee": nil},
							},
						},
					},
					{"$sort": bson.M{"project_tasks.task_end_date": 1}},
					{"$limit": 10},
					{
						"$project": bson.M{
							"project_name": "$address",
							"task_name":    "$project_tasks.task_name",
							"status":       "$project_tasks.task_status",
							"progress": bson.M{
								"$switch": bson.M{
									"branches": []bson.M{
										{"case": bson.M{"$eq": []interface{}{"$project_tasks.task_status", "Not Started"}}, "then": "0%"},
										{"case": bson.M{"$eq": []interface{}{"$project_tasks.task_status", "In Progress"}}, "then": "50%"},
										{"case": bson.M{"$eq": []interface{}{"$project_tasks.task_status", "review"}}, "then": "75%"},
										{"case": bson.M{"$eq": []interface{}{"$project_tasks.task_status", "Completed"}}, "then": "100%"},
									},
									"default": "0%",
								},
							},
							"project_end_date": "$project_tasks.task_end_date",
						},
					},
				},
			},
		},
	}
	return DashboardData
}
