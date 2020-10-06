package dynamodb

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

//Item struct for DynamoDb
type Item struct {
	CareerBuildFlag string
	BuildOnRun      bool
}

//ReadFromDynamoDB is reading "BuildOnRun" Key in Table "Career_Build_Table"
//It returns true if key value is "1" and returns false otherwise
func ReadFromDynamoDB(db *dynamodb.DynamoDB, tableName string) bool {

	result, err := db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"CareerBuildFlag": {
				S: aws.String("BuildSite"),
			},
		},
	})

	if err != nil {
		log.Fatalln(err.Error())
		return false
	}

	if result.Item == nil {
		msg := "Could not find the item"
		log.Fatalln(msg)
		return false
	}

	item := Item{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		log.Fatalln(err.Error())
	}

	return item.BuildOnRun

}

//UpdateDynamoDB will update "BuildOnKey" value to 0 in dynamo db table "Career_Build_Table"
func UpdateDynamoDB(db *dynamodb.DynamoDB, tableName string, val bool) {

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":Val": {
				BOOL: aws.Bool(val),
			},
		},
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"CareerBuildFlag": {
				S: aws.String("BuildSite"),
			},
		},
		//ReturnValues:     aws.String("UPDATED_NEW"),
		UpdateExpression: aws.String("set BuildOnRun = :Val"),
	}
	_, err := db.UpdateItem(input)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

}
