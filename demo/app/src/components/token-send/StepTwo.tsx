import { useEffect, useState } from "react";
import { QRCode } from "react-qrcode-logo";
import { getRoute, routeStart } from "../../api/router/router";
interface StepTwoProps {
  qrCodeString: string;
  session_id: string;
  uniqueStrings: string[];
  setUniqueStrings: (uniqueStrings: string[]) => void;
  goToStep: (step: number) => void;
}

export default function StepTwo({
  qrCodeString,
  session_id,
  uniqueStrings,
  setUniqueStrings,
  goToStep,
}: StepTwoProps) {
  const [loading, setLoading] = useState<boolean>(true);
  const [canContinue, setCanContinue] = useState(false);

  useEffect(() => {
    let intervalId = setInterval(getDevices, 3000);
    return () => clearInterval(intervalId);
  }, [session_id]);

  const getDevices = async () => {
    if (!session_id) return;
    try {
      const res = await getRoute(session_id);
      if (res.ok) {
        const data = await res.json();
        const uniqueSet = new Set(data);
        const uniqueArray: any = Array.from(uniqueSet);
        setUniqueStrings(uniqueArray); // Update state with unique strings
        if (uniqueArray.length > 1) {
          setLoading(false);
          uniqueArray.length >= process.env.REACT_APP_MINIMUM_DEVICES!
            ? setCanContinue(true)
            : setCanContinue(false);
        }
      } else {
        console.error("Failed to fetch data from the API");
        setLoading(false);
      }
    } catch (error) {
      console.error("Error:", error);
      setLoading(false);
    }
  };

  const sendStartApi = async () => {
    const res = await routeStart(session_id, uniqueStrings);
    if (res.status === 200) {
      goToStep(3);
    }
  };

  return (
    <>
      <div className="bg-white p-4 rounded-lg mx-8">
        <QRCode
          value={qrCodeString}
          size={250}
          bgColor={"#ffffff"}
          fgColor={"#0B51C6"}
          ecLevel="L"
          logoImage="/img/logo.svg"
          logoWidth={70}
          logoHeight={70}
          qrStyle="dots"
          eyeRadius={[
            {
              outer: [10, 10, 0, 10],
              inner: [0, 10, 10, 10],
            },
            [10, 10, 10, 0],
            [10, 0, 10, 10],
          ]}
        />
      </div>

      <div className="flex flex-col items-start justify-center ">
        <h2 className="text-white text-[30px] font-bold">
          Select Pairing Devices
        </h2>
        {loading ? (
          <p className="text-white">Looking for devices...</p>
        ) : uniqueStrings && uniqueStrings.length > 0 ? (
          <div className="flex my-4">
            {uniqueStrings.map((device, index) => (
              <div
                key={index}
                className="text-white mr-2 flex flex-col justify-center items-center border rounded-lg p-4"
              >
                <img
                  src="/img/device.svg"
                  width={46}
                  height={61}
                  alt="tablet"
                />
                <span className="text-xs mt-2">{device}</span>
              </div>
            ))}
          </div>
        ) : (
          ""
        )}

        <button
          onClick={sendStartApi}
          disabled={!canContinue}
          className="text-white px-10 py-4 me-5 disabled:text-gray-500 btn-custom my-2 mx-lg-2 my-sm-0 rounded-lg shadow-sm "
        >
          Continue
        </button>
      </div>
    </>
  );
}