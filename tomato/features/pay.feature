Feature: http feature example

    Scenario: Set and compare http-wiremock responses
        Given set "user-service" with path "/example" response code to 200 and response body
            """
                {"user_id":"Dutchessbramble"}
            """
        Then "httpclient" send request to "POST /pay" with body
            """
            {
                "amount":100, 
                "transaction_id":"abc123"
            }
            """
        And "httpclient" response code should be 200
        And "httpclient" response body should equal
            """
                 {
                    "data": {
                        "payment_id": 1
                    },
                    "status": "success"
                }
            """
        Then "db" table "payments" should look like
        | id | transaction_id | authorized_by   |
        | 1  | abc123         | Dutchessbramble |