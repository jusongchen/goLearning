package main

import (
	"github.com/go-ldap/ldap"
)

func main() {
req := ldap.SearchRequest {
type SearchRequest struct {
	BaseDN       :"internal.salesforce.com"
	Scope        :
	DerefAliases :
	SizeLimit    :
	TimeLimit    :
	TypesOnly    :
	Filter       :
	Attributes   :
	Controls     :
}
}
	
    
}


jusongch-ltm2.internal.salesforce.com

dcsfm1.internal.salesforce.com

ldapsearch -D "cn=directory manager" -W  -p 389 -h dcsfm1.internal.salesforce.com -b "dc=internal,dc=salesforce,dc=com" -s sub "(objectclass=*)"