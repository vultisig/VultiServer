import { useEffect, useState, useRef } from "react";
import { routeComplete } from "../api/router/router";
interface StepFourProps {
    devices?: string[];
    session_id?: string;
}

export default function StepFour({ devices = [], session_id = '' }: StepFourProps) {
    const [status, setStatus] = useState<string>('pending');
    const intervalIdRef = useRef<NodeJS.Timeout | null>(null);
    useEffect(() => {
        if (status !== 'done') intervalIdRef.current = setInterval(getCompleteDevice, 3000);
        return () => {
            if (intervalIdRef.current) clearInterval(intervalIdRef.current);
        };
    }, [session_id, status]);
    const getCompleteDevice = async () => {
        let lengthDevices = devices.length;
        const data = await (await routeComplete(session_id)).json();
        const uniqueDevices: any = Array.from(new Set(data));
        if (uniqueDevices.length == lengthDevices) {
            setStatus('done');
            if (intervalIdRef.current) {
                clearInterval(intervalIdRef.current);
            }
        }
    }
    return (
        <>
            <div className="w-full pt-20 pb-40 relative">
                {status == 'pending' ? (
                    <>
                        <div className="flex justify-center items-center mt-[-40px]">
                            <div className="loader" >
                                <span></span>
                                <span></span>
                                <span></span>
                                <span></span>
                            </div>
                            <div className="absolute text-white text-xs">
                                <span>Preparing Vault...</span>
                                <div className="flex justify-center mt-2">
                                    <div className="circle-1"></div>
                                    <div className="circle-2"></div>
                                </div>
                            </div>
                        </div>
                        <img className="bottom-0 left-0 right-0 text-white text-center text-lg absolute mx-auto" src="/img/anten.svg" width={40} height={40} alt="cellular" />
                        <p className="lg:w-[40%] w-full bottom-[-60px] left-0 right-0 mx-auto text-white text-center text-lg absolute">Keep devices connected to the internet
                            with VultiSig open</p>
                    </>

                ) : (
                    <div className="flex flex-col justify-center items-center">
                        <div className="done rounded-full w-[100px] h-[100px] flex flex-col justify-center items-center bg-[#33e6bf]">
                            <img src="img/checked.svg" />
                        </div>
                        <p className="text-white mt-4 text-xl">Done!</p>
                    </div>
                )}

            </div>

        </>
    )
}