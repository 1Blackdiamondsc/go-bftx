package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http" // Provides HTTP client and server implementations.
	"strconv"

	"os"

	"github.com/blockfreight/go-bftx/api/graphqlObj"
	apiHandler "github.com/blockfreight/go-bftx/api/handlers"
	"github.com/blockfreight/go-bftx/lib/app/bf_tx" // Provides some useful functions to work with LevelDB.
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	graphiql "github.com/mnmtanish/go-graphiql"
)

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	},
)

var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"getTransaction": &graphql.Field{
				Type: graphqlObj.TransactionType,
				Args: graphql.FieldConfigArgument{
					"Id": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					bftxID, isOK := p.Args["id"].(string)
					if !isOK {
						return nil, errors.New(strconv.Itoa(http.StatusBadRequest))
					}

					return apiHandler.GetTransaction(bftxID)
				},
			},
			"queryTransaction": &graphql.Field{
				Type: graphql.NewList(graphqlObj.TransactionType),
				Args: graphql.FieldConfigArgument{
					"Id": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					bftxID, isOK := p.Args["Id"].(string)
					if !isOK {
						return nil, errors.New(strconv.Itoa(http.StatusBadRequest))
					}

					return apiHandler.QueryTransaction(bftxID)
				},
			},
			"getInfo": &graphql.Field{
				Type: graphqlObj.InfoType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return apiHandler.GetInfo()
				},
			},
			"getTotal": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					log.Println("getTotal")
					return apiHandler.GetTotal()
				},
			},
		},
	})

var mutationType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"constructBFTX": &graphql.Field{
				Type: graphqlObj.TransactionType,
				Args: graphql.FieldConfigArgument{
					"Properties": &graphql.ArgumentConfig{
						Description: "Transaction properties.",
						Type:        graphqlObj.PropertiesInput,
					},
				},
				// Args: graphql.FieldConfigArgument{
				// 	"Properties": &graphql.ArgumentConfig{
				// 		Type: graphql.NewNonNull(graphql.Int),
				// 	},
				// },
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					log.Println("constructBFTX")
					bftx := bf_tx.BF_TX{}
					jsonProperties, err := json.Marshal(p.Args)
					if err = json.Unmarshal([]byte(jsonProperties), &bftx); err != nil {
						return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
					}

					return apiHandler.ConstructBfTx(bftx)
				},
			},
			"encryptBFTX": &graphql.Field{
				Type: graphqlObj.TransactionType,
				Args: graphql.FieldConfigArgument{
					"Id": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					bftxID, isOK := p.Args["Id"].(string)
					if !isOK {
						return nil, nil
					}

					return apiHandler.EncryptBFTX(bftxID)
				},
			},
			"decryptBFTX": &graphql.Field{
				Type: graphqlObj.TransactionType,
				Args: graphql.FieldConfigArgument{
					"Id": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					bftxID, isOK := p.Args["Id"].(string)
					if !isOK {
						return nil, nil
					}

					return apiHandler.DecryptBFTX(bftxID)
				},
			},
			"signBFTX": &graphql.Field{
				Type: graphqlObj.TransactionType,
				Args: graphql.FieldConfigArgument{
					"Id": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					bftxID, isOK := p.Args["Id"].(string)
					if !isOK {
						return nil, nil
					}

					return apiHandler.SignBfTx(bftxID)
				},
			},
			"broadcastBFTX": &graphql.Field{
				Type: graphqlObj.TransactionType,
				Args: graphql.FieldConfigArgument{
					"Id": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					bftxID, isOK := p.Args["Id"].(string)
					if !isOK {
						return nil, nil
					}

					return apiHandler.BroadcastBfTx(bftxID)
				},
			},
		},
	},
)

//Start start the API
func Start() error {
	http.HandleFunc("/bftx-api", httpHandler(&schema))
	http.HandleFunc("/", graphiql.ServeGraphiQL)
	http.HandleFunc("/graphql", serveGraphQL(schema))
	ex, err := os.Executable()
	if err != nil {
		fmt.Println(err)
	}
	file, err := os.Stat(ex)
	fmt.Println("Compiled: " + file.ModTime().String())
	fmt.Println("Now server is running on: http://localhost:8080")
	fmt.Println("Started")
	return http.ListenAndServe(":8080", nil)
}

func httpHandler(schema *graphql.Schema) func(http.ResponseWriter, *http.Request) {
	fmt.Println("In handler")
	return func(rw http.ResponseWriter, r *http.Request) {
		fmt.Println("In httpHandler")
		rw.Header().Set("Content-Type", "application/json")
		httpStatusResponse := http.StatusOK
		// parse http.Request into handler.RequestOptions
		opts := handler.NewRequestOptions(r)

		// inject context objects http.ResponseWrite and *http.Request into rootValue
		// there is an alternative example of using `net/context` to store context instead of using rootValue
		rootValue := map[string]interface{}{
			"response": rw,
			"request":  r,
			"viewer":   "john_doe",
		}

		// execute graphql query
		// here, we passed in Query, Variables and OperationName extracted from http.Request
		params := graphql.Params{
			Schema:         *schema,
			RequestString:  opts.Query,
			VariableValues: opts.Variables,
			OperationName:  opts.OperationName,
			RootObject:     rootValue,
		}
		fmt.Println(params.OperationName)
		fmt.Println("graphql.Do(params)")
		fmt.Println(params)
		result := graphql.Do(params)
		js, err := json.Marshal(result)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(rw, err.Error(), http.StatusInternalServerError)

		}
		if result.HasErrors() {
			httpStatusResponse, err = strconv.Atoi(result.Errors[0].Error())
			if err != nil {
				fmt.Println(err.Error())
				httpStatusResponse = http.StatusInternalServerError
			}
		}
		fmt.Println(httpStatusResponse)
		fmt.Println(js)
		rw.WriteHeader(httpStatusResponse)

		rw.Write(js)

	}

}

func serveGraphQL(s graphql.Schema) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("In serveGraphQL")
		sendError := func(err error) {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
		}

		req := &graphiql.Request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			sendError(err)
			return
		}

		res := graphql.Do(graphql.Params{
			Schema:        s,
			RequestString: req.Query,
		})

		if err := json.NewEncoder(w).Encode(res); err != nil {
			sendError(err)
		}
	}
}
