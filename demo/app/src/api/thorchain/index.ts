export const rpcEndpoint = "https://thornode.ninerealms.com";

export function getBalances(thorAddress: string): Promise<Response> {
  return fetch(`${rpcEndpoint}/cosmos/bank/v1beta1/balances/${thorAddress}`);
}

export function getAccountInfo(thorAddress: string): Promise<Response> {
  return fetch(`${rpcEndpoint}/auth/accounts/${thorAddress}`);
}

export async function broadcastSignedTransaction(
  signedTx: string
): Promise<any> {
  try {
    const response = await fetch(`${rpcEndpoint}/broadcast`, {
      method: "POST",
      body: signedTx,
    });

    if (response.status === 200) {
      const data = await response.json();
      console.log("Transaction broadcasted successfully:", data);
      return data;
    } else {
      console.error(
        "Failed to broadcast transaction:",
        response.status,
        response.statusText
      );
      return;
    }
  } catch (error) {
    console.error("Error broadcasting transaction:", error);
    return;
  }
}
