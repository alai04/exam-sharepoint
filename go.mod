module github.com/alai04/exam-sharepoint

go 1.16

require (
	github.com/joho/godotenv v1.4.0
	github.com/koltyakov/gosip v0.0.0-20220724185821-bc92c0983a11
	github.com/koltyakov/gosip-sandbox v0.0.0-20221012144252-38b6671ec9db
	github.com/piquette/finance-go v1.0.0
	github.com/spf13/viper v1.15.0
	github.com/xuri/excelize/v2 v2.7.0
)

require github.com/shopspring/decimal v1.3.1 // indirect

replace github.com/piquette/finance-go v1.0.0 => github.com/alai04/finance-go v1.0.1
