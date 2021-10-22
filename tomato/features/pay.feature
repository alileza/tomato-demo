Feature: http feature example

    Scenario: Pay endpoint should insert data to db correctly
        Given listen message from "rabbitmq" target "payments:created"
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
        And "httpclient" response body should contain
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
        Then message from "rabbitmq" target "payments:created" count should be 1
        Then message from "rabbitmq" target "payments:created" should contain
        """
           {"data":{"payment_id":1},"status":"success"}
        """

    Scenario:  Pay endpoint should failed to insert data when user service is down
        Given listen message from "rabbitmq" target "payments:created"
        Given set "user-service" with path "/example" response code to 500 and response body
            """
                {"error":"bad gateway"}
            """
        Then "httpclient" send request to "POST /pay" with body
            """
            {
                "amount":100, 
                "transaction_id":"abc123"
            }
            """
        And "httpclient" response code should be 500
        Then "db" table "payments" should look like
        | id | transaction_id | authorized_by   |
        Then message from "rabbitmq" target "payments:created" count should be 0
        