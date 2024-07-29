const vultisignerBaseUrl = process.env.REACT_APP_VULTISIGNER_BASE_URL;
const vultisigRelayer = process.env.REACT_APP_VULTISIG_RELAYER_URL;
export const endPoints = {
  // Router
  router: `${vultisigRelayer}router`,
  routerStart: `${vultisigRelayer}router/start`,
  routerComplete: `${vultisigRelayer}router/complete`,
  // Vault
  createVault: `${vultisignerBaseUrl}vault/create`,
  uploadVault: `${vultisignerBaseUrl}vault/upload`,
  downloadVault: `${vultisignerBaseUrl}vault/download`,
  sign: `${vultisignerBaseUrl}vault/sign`,
  getSignResult: `${vultisignerBaseUrl}vault/sign/response`,
};
