import { motion } from "framer-motion";
import { useState } from "react";
import { createVault } from "../api/create-vault/vault";

const variants = {
  initial: { x: "100%" },
  enter: { x: 0 },
  exit: { x: "-100%" },
};
interface StepTwoProps {
  goToStep: (step: number) => void;
  setSession_id: (session_id: string) => void;
  setQrCodeString: (qrCodeString: string) => void;
}
export default function StepTwo({
  goToStep,
  setSession_id,
  setQrCodeString,
}: StepTwoProps) {
  const [vaultName, setVaultName] = useState("");
  const [vaultPwd, setVaultPwd] = useState("");
  const [canContinue, setCanContinue] = useState(false);

  const CreateVault = async () => {
    const localPartyId = `VultiSignerApp-${
      Math.floor(Math.random() * 1000) + 1
    }`;
    const data = await (
      await createVault(vaultName, vaultPwd, localPartyId)
    ).json();
    setQrCodeString(
      `vultisig://vultisig.com?type=NewVault&tssType=Keygen&jsonData=${data.keygen_msg}`
    );
    setSession_id(data.session_id);
    goToStep(3);
  };
  const checkContiune = (input: string) => {
    if (input.length > 0) {
      setCanContinue(true);
    } else {
      setCanContinue(false);
    }
  };
  return (
    <>
      <motion.div
        initial="initial"
        animate="enter"
        exit="exit"
        variants={variants}
        transition={{ duration: 0.3 }}
      >
        <h1 className="text-white text-[60px] font-bold">
          Vulti <span className="color-custom">Signer</span>
        </h1>
        <input
          type="text"
          onChange={(e) => {
            setVaultName(e.target.value);
            checkContiune(e.target.value);
          }}
          className="bg-[#11243E] text-white w-full my-4 p-3 rounded-lg"
          placeholder="Vault Name"
        />
        <input
          type="password"
          onChange={(e) => {
            setVaultPwd(e.target.value);
            checkContiune(e.target.value);
          }}
          className="bg-[#11243E] text-white w-full my-4 p-3 rounded-lg"
          placeholder="Vault Password"
        />
        <button
          onClick={CreateVault}
          disabled={!canContinue}
          className="text-white px-10 py-4 me-5  btn-custom my-2 mx-lg-2 my-sm-0 rounded-lg shadow-sm"
        >
          Create Vault
        </button>
      </motion.div>
    </>
  );
}
