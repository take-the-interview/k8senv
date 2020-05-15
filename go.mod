module github.com/take-the-interview/k8senv

go 1.14

require (
	github.com/aws/aws-sdk-go v1.30.28
	github.com/take-the-interview/shep v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37 // indirect
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20200515095857-1151b9dac4a9 // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
)

replace github.com/take-the-interview/shep => ../shep
