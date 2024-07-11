const variants = {
    initial: { x: "100%" },
    enter: { x: 0 },
    exit: { x: "-100%" },
};

interface StepOneProps {
    goToStep: (step: number) => void;
}
export default function ({ goToStep }: StepOneProps) {

    return (
        <>
            <h1 className="text-white text-[60px] font-bold">Vulti <span className="color-custom">Signer</span></h1>
            <h2 className="my-8 text-white text-[22px] font-bold">SECURE CRYPTO VAULT</h2>
            <button onClick={() => goToStep(2)}
                className="text-white px-8 py-4 me-5  btn-custom my-2 mx-lg-2 my-sm-0 rounded-lg shadow-sm">
                Create a New Vault
            </button>
        </>
    )
}