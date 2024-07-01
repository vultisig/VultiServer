"use client"

import StepFour from "../../components/step-four";
import StepOne from "../../components/step-one"
import StepThree from "../../components/step-three";
import StepTwo from "../../components/step-two"
import { useState } from "react"
import { AnimatePresence } from "framer-motion";
export default function Signer() {
  const [currentStep, setCurrentStep] = useState<number>(1);
  const [uniqueStrings, setUniqueStrings] = useState<string[]>([]);
  const [qrCodeString, setQrCodeString] = useState<string>("");
  const [session_id, setSession_id] = useState<string>("");

  const goToStep = (step: number) => {
    setCurrentStep(step);
  };

  return (
    <>
      <section className="h-[70vh] flex justify-center items-center">
        {(currentStep === 1 || currentStep === 2) && (
          <AnimatePresence mode='wait'>
            <>
              <div className="mx-auto w-full mt-5 rounded-xl bg-[#061B3A] max-w-screen-xl px-4 py-6 lg:py-20 flex justify-between overflow-hidden">
                <div className="lg:w-[70%] w-full flex flex-wrap justify-center lg:justify-between mx-auto">
                  <img src={`/img/main-logo.svg`}
                    width={300} height={300} className="object-contain" alt="logo" />
                  <div className="flex flex-col items-start justify-center flex-wrap">
                    {currentStep === 1 && <StepOne goToStep={goToStep} />}
                    {currentStep === 2 && <StepTwo setSession_id={setSession_id} goToStep={goToStep} setQrCodeString={setQrCodeString} />}
                  </div>
                </div>
              </div>
            </>
          </AnimatePresence>
        )
        }
        {currentStep === 3 && (
          <StepThree
            qrCodeString={qrCodeString}
            session_id={session_id}
            uniqueStrings={uniqueStrings}
            setUniqueStrings={setUniqueStrings}
            goToStep={goToStep}
          />
        )}
        {currentStep === 4 && <StepFour session_id={session_id} devices={uniqueStrings} />}
      </section>
    </>
  )
}