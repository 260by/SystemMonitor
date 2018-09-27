package ssh

import (
	"golang.org/x/crypto/ssh"
	// "io/ioutil"
	"time"
	"strconv"
	// "fmt"
)

// Connect ssh连接，返回ssh client, authentication长度大于100判断为使用SSH私钥认证
func Connect(user, host string, port int, authentication string) (client *ssh.Client, err error) {
	auth := make([]ssh.AuthMethod,0)

	if len(authentication) > 100 {
		signer, err := ssh.ParsePrivateKey([]byte(authentication))
		if err != nil {
			return nil, err
		}
		// fmt.Printf("Public Key Used. %s\n", host)
		auth = append(auth, ssh.PublicKeys(signer))
	} else {
		// fmt.Printf("Password Used. %s\n", host)
		auth = append(auth, ssh.Password(authentication))
	}
	

	clientConfig := &ssh.ClientConfig{
		User:            user,
		Auth:    	     auth,
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := host + ":" + strconv.Itoa(port)

	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	return client, nil
}
