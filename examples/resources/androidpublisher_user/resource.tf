resource "androidpublisher_user" "test" {
  email                         = "my-service@myproject-123456.iam.gserviceaccount.com"
  developer_id                  = "1234567891234567891"
  developer_account_permissions = ["CAN_VIEW_APP_QUALITY_GLOBAL"]
}