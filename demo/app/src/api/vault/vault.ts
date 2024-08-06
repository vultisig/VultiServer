import { getAuthHeader } from "../../utils/utils";
import { endPoints } from "../endpoints";

export function createVault(
  vaultName: string,
  password: string,
  localPartyId: string
): Promise<Response> {
  return fetch(endPoints.createVault, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: getAuthHeader(),
    },
    body: JSON.stringify({
      name: vaultName,
      local_party_id: localPartyId,
      encryption_password: password,
    }),
  });
}

export function getVault(
  vaultPublicKeyEcdsa: string,
  passwd: string
): Promise<Response> {
  return fetch(`${endPoints.getVault}/${vaultPublicKeyEcdsa}`, {
    headers: {
      "x-password": passwd,
    },
  });
}

export function signMessages(passwd: string, data: string): Promise<Response> {
  return fetch(endPoints.sign, {
    method: "POST",
    headers: {
      "x-password": passwd,
      "Content-Type": "application/json",
    },
    body: data,
  });
}

export function getSignResult(taskId: string): Promise<Response> {
  return fetch(`${endPoints.getSignResult}/${taskId}`);
}
