package abb

import(	
	"os"	
	"strconv"
)

func Envint(key string, defaultvalue int) int{
	valuestr, haskey := os.LookupEnv(key)
	if haskey{
		intvalue, err := strconv.Atoi(valuestr)
		if err != nil{
			return defaultvalue
		}else{
			return intvalue
		}
	}
	return defaultvalue
}

func Envstr(key string, defaultvalue string) string{
	valuestr, haskey := os.LookupEnv(key)
	if haskey{
		return valuestr
	}
	return defaultvalue
}
