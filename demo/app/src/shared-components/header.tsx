

import {  Link } from "react-router-dom";

export default function Header() {
    return (
        <header className="">
            <div className="mx-auto w-full mt-5 rounded-xl bg-[#061B3A] max-w-screen-xl px-4 py-6 lg:py-5 flex justify-between">
                <div className="flex items-center text-white">
                    <img width={50} height={50} src={`/img/logo.svg`} alt="logo vultisigner" />
                    <h2 className="text-xl mx-4 font-bold">VultiSigner</h2>
                </div>
                <a href={'#'}
                    className="text-white px-8 py-2 me-5  btn-custom my-2 mx-lg-2 my-sm-0 rounded-lg shadow-sm">
                    Demo
                </a>
            </div>
           
        </header>
    );
}
