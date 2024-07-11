import { encodeToBase64 } from "../../utils/utils";
import { endPoints } from "../endpoints";
const userName = process.env.REACT_APP_VULTISIGNER_USER;
const passWord = process.env.REACT_APP_VULTISIGNER_PASSWORD;
export function createVault(vaultName: string, password: string): Promise<Response> {
  let userPass = `${userName}:${passWord}`
  return fetch(endPoints.createVault, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      "Authorization": `Basic ${encodeToBase64(userPass)}`
    },
    body: JSON.stringify({
      name: vaultName,
      encryption_password: password,
    }),
  });
}