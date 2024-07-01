generate-demo:
	yes | rm -rf demo/generated/* && cd demo/app && npm i && REACT_APP_VULTISIGNER_BASE_URL="http://127.0.0.1:8080/" REACT_APP_VULTISIG_RELAYER_URL="https://api.vultisig.com/" REACT_APP_MINIMUM_DEVICES=2  REACT_APP_VULTISIGNER_USER="username" REACT_APP_VULTISIGNER_PASSWORD="password"  npm run build && mv build/* ../generated
