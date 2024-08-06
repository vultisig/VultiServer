interface StepThreeProps {
  status: string;
}

export default function StepThree({ status }: StepThreeProps) {
  return (
    <>
      <div className="w-full mt-20 pb-20 mb-8 pb-40 relative">
        {status === "pending" ? (
          <>
            <div className="flex justify-center items-center mt-[-40px]">
              <div className="loader">
                <span></span>
                <span></span>
                <span></span>
                <span></span>
              </div>
              <div className="absolute text-white text-xs">
                <span>Signing...</span>
                <div className="flex justify-center mt-2">
                  <div className="circle-1"></div>
                  <div className="circle-2"></div>
                </div>
              </div>
            </div>
            <img
              className="bottom-0 left-0 right-0 text-white text-center text-lg absolute mx-auto"
              src="/img/anten.svg"
              width={40}
              height={40}
              alt="cellular"
            />
            <p className="lg:w-[40%] w-full bottom-[-60px] left-0 right-0 mx-auto text-white text-center text-lg absolute">
              Keep devices connected to the internet with VultiSig open
            </p>
          </>
        ) : (
          <div className="flex flex-col justify-center items-center">
            <div className="done rounded-full w-[100px] h-[100px] flex flex-col justify-center items-center bg-[#33e6bf]">
              <img src="img/checked.svg" alt="checked" />
            </div>
            <p className="text-white mt-4 text-xl">Done!</p>
          </div>
        )}
      </div>
    </>
  );
}
