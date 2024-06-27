
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
export default function StepTwo({ goToStep, setSession_id, setQrCodeString }: StepTwoProps) {
    const [vaultName, setVaultName] = useState('');
    const [canContinue, setCanContinue] = useState(false)
    const generateRandomString = (): string => {
        const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
        let result = '';
        for (let i = 0; i < 10; i++) {
            result += chars.charAt(Math.floor(Math.random() * chars.length));
        }
        return result;
    };
    const CreateVault = async () => {
        const passNew = generateRandomString();
        const data = await (await createVault(vaultName, passNew)).json()
        setQrCodeString(`vultisig://vultisig.com?type=NewVault&tssType=Keygen&jsonData={"Keygen":{"_0":{"encryptionKeyHex":"${data.hex_encryption_key}","hexChainCode":"${data.hex_chain_code}","serviceName":"VultiSigner-001",
                    "sessionID":"${data.session_id}","useVultisigRelay":true,"vaultName":"${vaultName}"}}}`)
        setSession_id(data.session_id)
        goToStep(3);
    }
    const checkContiune = (input: string) => {
        if (input.length > 0) {
            setCanContinue(true)
        } else {
            setCanContinue(false)
        }
    }
    return (
        <>
            <motion.div
                initial="initial"
                animate="enter"
                exit="exit"
                variants={variants}
                transition={{ duration: 0.3 }}
            >
                <h1 className="text-white text-[60px] font-bold">Vulti <span className="color-custom">Signer</span></h1>
                <input type="text" onChange={(e) => { setVaultName(e.target.value); checkContiune(e.target.value) }}
                    className="bg-[#11243E] text-white w-full my-4 p-3 rounded-lg" placeholder="Vault Name" />

                <button onClick={CreateVault} disabled={!canContinue}
                    className="text-white px-10 py-4 me-5  btn-custom my-2 mx-lg-2 my-sm-0 rounded-lg shadow-sm">
                    Create Vault
                </button>
            </motion.div>
        </>
    )
}