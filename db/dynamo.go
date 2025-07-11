package db

import (
	"context"
	"errors"
	"fmt"

	paginate "github.com/techvuya/vuya-go-utils/paginate"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"

	"github.com/aws/aws-sdk-go-v2/config"
)

var ErrQueryNoData = errors.New("ErrQueryNoData")

type DynamoDatabaseClient struct {
	dbEnvPrefix  string
	dynamoClient *dynamodb.Client
}

func CreateDynamoDatabaseClient(awsSessionRegion, dbEnvPrefix string) (*DynamoDatabaseClient, error) {
	dynamoClient, err := generateNewDynamoAccessSession(awsSessionRegion)
	if err != nil {
		return nil, err
	}
	return &DynamoDatabaseClient{
		dynamoClient: dynamoClient,
		dbEnvPrefix:  dbEnvPrefix,
	}, nil
}

func generateNewDynamoAccessSession(awsSessionRegion string) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
		o.Region = awsSessionRegion
		return nil
	})
	if err != nil {
		return nil, err
	}
	awsv2.AWSV2Instrumentor(&cfg.APIOptions)
	svc := dynamodb.NewFromConfig(cfg)
	return svc, nil
}

func (c DynamoDatabaseClient) GetTableUrl(tableName string) string {
	if c.dbEnvPrefix == "" {
		return tableName
	}

	return c.dbEnvPrefix + "." + tableName
}

func (c DynamoDatabaseClient) Get(ctx context.Context, tableName string, keys map[string]types.AttributeValue, resultDataPointer interface{}) error {
	tableUrl := c.GetTableUrl(tableName)
	result, err := c.dynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:              aws.String(tableUrl),
		Key:                    keys,
		ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
	})
	if err != nil {
		return err
	}
	if result.Item == nil {
		return ErrQueryNoData
	}

	err = attributevalue.UnmarshalMap(result.Item, &resultDataPointer)
	if err != nil {
		return err
	}
	return nil
}
func (c DynamoDatabaseClient) GetBatch(ctx context.Context, tableName string, keys []map[string]types.AttributeValue, resultDataPointer interface{}) error {
	tableUrl := c.GetTableUrl(tableName)
	request := map[string]types.KeysAndAttributes{
		tableUrl: {
			Keys: keys,
		},
	}
	result, err := c.dynamoClient.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
		RequestItems: request,
	})
	if err != nil {
		return err
	}
	if result.Responses[tableUrl] == nil {
		return ErrQueryNoData
	}

	err = attributevalue.UnmarshalListOfMaps(result.Responses[tableUrl], &resultDataPointer)
	if err != nil {
		return err
	}
	return nil
}

func GetResponseCursor(lastEvaluatedKey map[string]types.AttributeValue, key string) (string, error) {
	atributeValuex := lastEvaluatedKey[key]
	var eventIdString string
	err := attributevalue.Unmarshal(atributeValuex, &eventIdString)
	if err != nil {
		return "", err
	}
	return eventIdString, nil
}

func getResponseCursor(lastEvaluatedKey map[string]types.AttributeValue, key string) (string, error) {
	atributeValuex := lastEvaluatedKey[key]
	var eventIdString string
	err := attributevalue.Unmarshal(atributeValuex, &eventIdString)
	if err != nil {
		return "", err
	}
	return eventIdString, nil
}

func CreateDynamoPaginateRequest(keyName, keyValue, cursorName, cursorValue, order string) (expression.Expression, error) {
	var keyEx expression.KeyConditionBuilder

	if cursorValue == "" {
		keyEx = expression.Key(keyName).Equal(expression.Value(keyValue))
	} else if order == "DESC" {
		keyEx = expression.Key(keyName).Equal(expression.Value(keyValue)).And(expression.Key(cursorName).LessThan(expression.Value(cursorValue)))
	} else {
		keyEx = expression.Key(keyName).Equal(expression.Value(keyValue)).And(expression.Key(cursorName).GreaterThan(expression.Value(cursorValue)))
	}
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		return expression.Expression{}, err
	}
	return expr, nil
}

func CreateDynamoPaginateRequestWithCondition(keyName, keyValue, cursorName, cursorValue, order string, filterCond expression.ConditionBuilder) (expression.Expression, error) {
	var keyEx expression.KeyConditionBuilder

	if cursorValue == "" {
		keyEx = expression.Key(keyName).Equal(expression.Value(keyValue))
	} else if order == "DESC" {
		keyEx = expression.Key(keyName).Equal(expression.Value(keyValue)).And(expression.Key(cursorName).LessThan(expression.Value(cursorValue)))
	} else {
		keyEx = expression.Key(keyName).Equal(expression.Value(keyValue)).And(expression.Key(cursorName).GreaterThan(expression.Value(cursorValue)))
	}
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).WithFilter(filterCond).Build()
	if err != nil {
		return expression.Expression{}, err
	}
	return expr, nil
}
func (c DynamoDatabaseClient) Query(
	ctx context.Context,
	tableName string,
	index string,
	expr expression.Expression,
	limitItems *int32,
	resultDataPointer interface{},
	cursorKey string, paginateParams paginate.AgPaginateOptionsRequest) (string, error) {
	scanIndexForward := false
	if paginateParams.GetOrder() == "ASC" {
		scanIndexForward = true
	}

	tableUrl := c.GetTableUrl(tableName)

	queryParams := dynamodb.QueryInput{
		TableName:                 aws.String(tableUrl),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
		Limit:                     limitItems,
		ScanIndexForward:          &scanIndexForward,
		ReturnConsumedCapacity:    types.ReturnConsumedCapacityTotal,
	}

	if index != "" {
		queryParams.IndexName = aws.String(index)
	}

	result, err := c.dynamoClient.Query(ctx, &queryParams)
	if err != nil {
		return "", err
	}

	if result.Items == nil || len(result.Items) == 0 {
		return "", ErrQueryNoData
	}

	cursorLastKey := ""
	if result.LastEvaluatedKey != nil && cursorKey != "" {
		cursorLastKey, err = getResponseCursor(result.LastEvaluatedKey, cursorKey)
		if err != nil {
			return "", err
		}
	}

	err = attributevalue.UnmarshalListOfMaps(result.Items, &resultDataPointer)
	if err != nil {
		return "", err
	}
	return cursorLastKey, nil
}

type CursorItem interface {
	GetCursorID() string
}

func (c DynamoDatabaseClient) QueryPaginate(
	ctx context.Context,
	tableName string,
	index string,
	expr expression.Expression,
	limitItems int32,
	resultDataPointer *[]CursorItem,
	cursorKey string, paginateParams paginate.AgPaginateOptionsRequest) (string, error) {
	scanIndexForward := false
	if paginateParams.GetOrder() == "ASC" {
		scanIndexForward = true
	}
	tableUrl := c.GetTableUrl(tableName)

	limitItemsFormat := limitItems + 1

	queryParams := dynamodb.QueryInput{
		TableName:                 aws.String(tableUrl),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
		Limit:                     &limitItemsFormat,
		ScanIndexForward:          &scanIndexForward,
		ReturnConsumedCapacity:    types.ReturnConsumedCapacityTotal,
	}

	if index != "" {
		queryParams.IndexName = aws.String(index)
	}

	result, err := c.dynamoClient.Query(ctx, &queryParams)
	if err != nil {
		return "", err
	}

	if result.Items == nil || len(result.Items) == 0 {
		return "", ErrQueryNoData
	}

	cursorLastKey := ""
	// if result.LastEvaluatedKey != nil && cursorKey != "" {
	// 	cursorLastKey, err = getResponseCursor(result.LastEvaluatedKey, cursorKey)
	// 	if err != nil {
	// 		return "", err
	// 	}
	// }

	err = attributevalue.UnmarshalListOfMaps(result.Items, &resultDataPointer)
	if err != nil {
		return "", err
	}
	responseSize := len(*resultDataPointer)

	if int32(responseSize) > limitItems {
		*resultDataPointer = (*resultDataPointer)[:limitItems]
		cursorLastKey = (*resultDataPointer)[len(*resultDataPointer)-1].GetCursorID()
	} else {
		cursorLastKey = ""
	}

	return cursorLastKey, nil
}

func QueryPaginate[T CursorItem](
	ctx context.Context,
	dbClient *DynamoDatabaseClient,
	tableName string,
	index string,
	expr expression.Expression,
	limitItems int32,
	resultDataPointer *[]T,
	cursorKey string, paginateParams paginate.AgPaginateOptionsRequest) (string, error) {
	scanIndexForward := false
	if paginateParams.GetOrder() == "ASC" {
		scanIndexForward = true
	}

	limitItemsFormat := limitItems + 1

	tableUrl := dbClient.GetTableUrl(tableName)

	queryParams := dynamodb.QueryInput{
		TableName:                 aws.String(tableUrl),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
		Limit:                     &limitItemsFormat,
		ScanIndexForward:          &scanIndexForward,
		ReturnConsumedCapacity:    types.ReturnConsumedCapacityTotal,
	}

	if index != "" {
		queryParams.IndexName = aws.String(index)
	}

	result, err := dbClient.dynamoClient.Query(ctx, &queryParams)
	if err != nil {
		return "", err
	}

	if result.Items == nil || len(result.Items) == 0 {
		return "", ErrQueryNoData
	}

	cursorLastKey := ""
	// if result.LastEvaluatedKey != nil && cursorKey != "" {
	// 	cursorLastKey, err = getResponseCursor(result.LastEvaluatedKey, cursorKey)
	// 	if err != nil {
	// 		return "", err
	// 	}
	// }

	err = attributevalue.UnmarshalListOfMaps(result.Items, &resultDataPointer)
	if err != nil {
		return "", err
	}
	responseSize := len(*resultDataPointer)

	if int32(responseSize) > limitItems {
		*resultDataPointer = (*resultDataPointer)[:limitItems]
		cursorLastKey = (*resultDataPointer)[len(*resultDataPointer)-1].GetCursorID()
	} else {
		cursorLastKey = ""
	}

	return cursorLastKey, nil
}

func (c DynamoDatabaseClient) QueryAllItems(
	ctx context.Context,
	tableName string,
	index string,
	expr expression.Expression,
	resultDataPointer interface{}) error {

	tableUrl := c.GetTableUrl(tableName)

	queryParams := dynamodb.QueryInput{
		TableName:                 aws.String(tableUrl),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ReturnConsumedCapacity:    types.ReturnConsumedCapacityTotal,
	}

	if index != "" {
		queryParams.IndexName = aws.String(index)
	}

	var allItems []map[string]types.AttributeValue

	paginator := dynamodb.NewQueryPaginator(c.dynamoClient, &queryParams)

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}

		allItems = append(allItems, output.Items...)
	}

	if len(allItems) == 0 {
		fmt.Println("@ErrorNoItemsQueryAll")
		return nil
	}

	err := attributevalue.UnmarshalListOfMaps(allItems, resultDataPointer)
	if err != nil {
		return err
	}

	return nil
}

func (c DynamoDatabaseClient) QueryCount(
	ctx context.Context,
	tableName string,
	index string,
	expr expression.Expression) (int32, error) {
	var count int32 = 0
	tableUrl := c.GetTableUrl(tableName)
	var lastEvaluatedKey map[string]types.AttributeValue
	var limitItems int32 = 5
	var consumedCapacity float64 = 0
	for {
		queryParams := dynamodb.QueryInput{
			TableName:                 aws.String(tableUrl),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			KeyConditionExpression:    expr.KeyCondition(),
			ReturnConsumedCapacity:    types.ReturnConsumedCapacityTotal,
			Limit:                     &limitItems,
		}

		if index != "" {
			queryParams.IndexName = aws.String(index)
		}

		if lastEvaluatedKey != nil {
			queryParams.ExclusiveStartKey = lastEvaluatedKey
		}

		result, err := c.dynamoClient.Query(ctx, &queryParams)
		if err != nil {
			return 0, err
		}

		consumedCapacity += *result.ConsumedCapacity.CapacityUnits

		count += result.Count

		lastEvaluatedKey = result.LastEvaluatedKey
		if lastEvaluatedKey == nil {
			break
		}
	}

	return count, nil
}

func (c DynamoDatabaseClient) QueryBatchItems(
	ctx context.Context,
	tableName string,
	index string,
	keys []map[string]types.AttributeValue,
	resultDataPointer interface{}) error {
	tableUrl := c.GetTableUrl(tableName)
	queryParams := dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			tableUrl: {
				Keys: keys,
			},
		},
		ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
	}

	result, err := c.dynamoClient.BatchGetItem(ctx, &queryParams)
	if err != nil {
		return err
	}

	if result.Responses == nil || len(result.Responses[tableUrl]) == 0 {
		return ErrQueryNoData
	}

	err = attributevalue.UnmarshalListOfMaps(result.Responses[tableUrl], resultDataPointer)
	if err != nil {
		return err
	}
	return nil
}

func (c DynamoDatabaseClient) QueryOne(ctx context.Context, tableName, index string, expr expression.Expression, resultDataPointer interface{}) error {
	limitItems := int32(1)
	tableUrl := c.GetTableUrl(tableName)
	queryParams := &dynamodb.QueryInput{
		TableName:                 aws.String(tableUrl),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		Limit:                     &limitItems,
		ReturnConsumedCapacity:    types.ReturnConsumedCapacityTotal,
	}

	if index != "" {
		queryParams.IndexName = aws.String(index)
	}
	result, err := c.dynamoClient.Query(ctx, queryParams)
	if err != nil {
		return err
	}
	if result.Items == nil || len(result.Items) == 0 {
		return ErrQueryNoData
	}

	err = attributevalue.UnmarshalListOfMaps(result.Items, &resultDataPointer)
	if err != nil {
		return err
	}
	return nil
}

func (c DynamoDatabaseClient) UpdateItem(
	ctx context.Context,
	tableName string,
	key map[string]types.AttributeValue,
	updateExpression string, expressionAttributeValues map[string]types.AttributeValue, conditionExpression string) error {
	tableUrl := c.GetTableUrl(tableName)
	input := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(tableUrl),
		Key:                       key,
		UpdateExpression:          aws.String(updateExpression),
		ExpressionAttributeValues: expressionAttributeValues,
		ReturnValues:              types.ReturnValueAllNew,
	}
	if conditionExpression != "" {
		input.ConditionExpression = aws.String(conditionExpression)
	}
	_, err := c.dynamoClient.UpdateItem(ctx, input)
	if err != nil {
		return err
	}
	return nil
}

func (x *NoSqlTransaction) AddTransactionUpdateQuery(ctx context.Context, tableName string, key map[string]types.AttributeValue,
	updateExpression string, expressionAttributeValues map[string]types.AttributeValue, conditionExpression string) error {
	tableUrl := x.GetTableUrl(tableName)
	input := &types.Update{
		TableName:                 aws.String(tableUrl),
		Key:                       key,
		UpdateExpression:          aws.String(updateExpression),
		ExpressionAttributeValues: expressionAttributeValues,
	}
	if conditionExpression != "" {
		input.ConditionExpression = aws.String(conditionExpression)
	}
	updateTx := types.TransactWriteItem{
		Update: input,
	}
	x.items = append(x.items, updateTx)
	return nil
}

func (c DynamoDatabaseClient) PutItem(ctx context.Context, tableName string, data interface{}) error {
	tableUrl := c.GetTableUrl(tableName)
	av, err := attributevalue.MarshalMap(data)
	if err != nil {
		return err
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableUrl),
	}
	_, err = c.dynamoClient.PutItem(ctx, input)
	if err != nil {
		return err
	}
	return nil
}

type NoSqlTransaction struct {
	items       []types.TransactWriteItem
	dbEnvPrefix string
}

func (c NoSqlTransaction) GetTableUrl(tableName string) string {
	if c.dbEnvPrefix == "" {
		return tableName
	}

	return c.dbEnvPrefix + "." + tableName
}
func CreateNoSqlTransaction(dynamoClient *DynamoDatabaseClient) *NoSqlTransaction {
	return &NoSqlTransaction{
		items:       []types.TransactWriteItem{},
		dbEnvPrefix: dynamoClient.dbEnvPrefix,
	}
}

func (x *NoSqlTransaction) AddTransaction(transaction types.TransactWriteItem) {
	x.items = append(x.items, transaction)
}

func (x *NoSqlTransaction) AddTransactionPut(tableName string, data interface{}) error {
	tableUrl := x.GetTableUrl(tableName)
	dataRaw, err := attributevalue.MarshalMap(data)
	if err != nil {
		return err
	}
	putTx := types.TransactWriteItem{
		Put: &types.Put{
			TableName: aws.String(tableUrl),
			Item:      dataRaw,
		},
	}
	x.items = append(x.items, putTx)
	return nil
}

func (x *NoSqlTransaction) AddTransactionPutExpr(tableName string, data interface{}, expr expression.Expression) error {
	tableUrl := x.GetTableUrl(tableName)
	dataRaw, err := attributevalue.MarshalMap(data)
	if err != nil {
		return err
	}
	putTx := types.TransactWriteItem{
		Put: &types.Put{
			TableName:           aws.String(tableUrl),
			Item:                dataRaw,
			ConditionExpression: expr.Condition(),
		},
	}
	x.items = append(x.items, putTx)
	return nil
}
func (x *NoSqlTransaction) AddTransactionDelete(tableName string, key map[string]types.AttributeValue) error {
	tableUrl := x.GetTableUrl(tableName)
	putTx := types.TransactWriteItem{
		Delete: &types.Delete{
			TableName: aws.String(tableUrl),
			Key:       key,
		},
	}
	x.items = append(x.items, putTx)
	return nil
}

func (x *NoSqlTransaction) BuildTransaction() *dynamodb.TransactWriteItemsInput {
	return &dynamodb.TransactWriteItemsInput{
		TransactItems: x.items,
	}
}

func (c DynamoDatabaseClient) ExecuteTransaction(ctx context.Context, params *NoSqlTransaction) error {
	response, err := c.dynamoClient.TransactWriteItems(ctx, params.BuildTransaction())
	if err != nil {
		return err
	}
	fmt.Println("-----------------------")
	fmt.Println(response)
	fmt.Println("-----------------------")
	return nil
}

func (x *NoSqlTransaction) AddTransactionUpdate(ctx context.Context, tableName string, key map[string]types.AttributeValue, expr expression.Expression) error {
	tableUrl := x.GetTableUrl(tableName)
	updateItem := &types.Update{
		TableName:                 aws.String(tableUrl),
		Key:                       key,
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ConditionExpression:       expr.Condition(),
	}
	updateTx := types.TransactWriteItem{
		Update: updateItem,
	}
	x.items = append(x.items, updateTx)
	return nil
}

type SelectCountParams struct {
	TableName        string
	IndexName        *string                     // Optional, use nil for primary index
	FilterExpression expression.ConditionBuilder // Optional, use expression.ConditionBuilder{} if no filter is needed
}

// selectCount performs a count operation on a DynamoDB table or index
func (c DynamoDatabaseClient) SelectCount(ctx context.Context, params SelectCountParams) (int64, error) {
	var count int64
	var lastEvaluatedKey map[string]types.AttributeValue

	for {
		builder := expression.NewBuilder()

		if !params.FilterExpression.IsSet() {
			builder = builder.WithFilter(params.FilterExpression)
		}

		expr, err := builder.Build()
		if err != nil {
			return 0, fmt.Errorf("failed to build expression: %w", err)
		}

		input := &dynamodb.ScanInput{
			TableName:                 aws.String(params.TableName),
			IndexName:                 params.IndexName,
			Select:                    types.SelectCount,
			ConsistentRead:            aws.Bool(false),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			FilterExpression:          expr.Filter(),
		}

		if lastEvaluatedKey != nil {
			input.ExclusiveStartKey = lastEvaluatedKey
		}

		result, err := c.dynamoClient.Scan(ctx, input)
		if err != nil {
			return 0, fmt.Errorf("failed to scan DynamoDB table: %w", err)
		}

		count += int64(result.Count)

		lastEvaluatedKey = result.LastEvaluatedKey
		if lastEvaluatedKey == nil {
			break
		}
	}

	return count, nil
}
