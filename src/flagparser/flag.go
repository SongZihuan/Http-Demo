package flagparser

import (
	"flag"
	"fmt"
)

func InitFlag() (err error) {
	defer func() {
		e := recover()
		if e != nil {
			err = fmt.Errorf("%v", e)
			return
		}
	}()

	flag.StringVar(&HttpAddress, "address", HttpAddress, "http server listen address")
	flag.StringVar(&HttpAddress, "http-address", HttpAddress, "http server listen address")

	flag.StringVar(&HttpsAddress, "https-address", HttpsAddress, "https server listen address")
	flag.StringVar(&HttpsDomain, "https-domain", HttpsDomain, "https server domain")
	flag.StringVar(&HttpsEmail, "https-email", HttpsEmail, "https cert email")
	flag.StringVar(&HttpsCertDir, "https-cert-dir", HttpsCertDir, "https cert save dir")

	flag.StringVar(&HttpsAliyunKey, "https-aliyun-dns-access-key", HttpsAliyunKey, "aliyun access key")
	flag.StringVar(&HttpsAliyunSecret, "https-aliyun-dns-access-secret", HttpsAliyunSecret, "aliyun access secret")

	flag.BoolVar(&DryRun, "dry-run", DryRun, "only parser the options")

	flag.BoolVar(&Verbose, "version", Verbose, "show the version")
	flag.BoolVar(&Verbose, "v", Verbose, "show the version")

	flag.Parse()

	return nil
}
