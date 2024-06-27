

export default function Footer() {
    return (
  <footer>
    <div className="mx-auto w-full max-w-screen-xl p-4 py-6 lg:py-8 border-t border-color">
        <div className="md:flex md:justify-between">
          <div className="mb-6 md:mb-0">
              <a href="#" className="flex items-center">
              <img width={50} height={50}  src={`/img/logo.svg`} alt="logo vultisigner" />
              <h2 className="text-xl mx-4 font-bold text-white">VultiSigner</h2>
              </a>
              <div className="flex mt-8">
          <img width={50} height={50} className="mx-2"  src={`/img/twitter.svg`} alt="logo vultisigner" />
          <img width={50} height={50} className="mx-2" src={`/img/github-sign.svg`} alt="logo vultisigner" />
          </div>
          </div>
          
          <div className="grid grid-cols-2 gap-8 sm:gap-6 sm:grid-cols-3">
              <div>
                  <h2 className="mb-6 text-sm font-semibold text-gray-900 uppercase dark:text-white">Support</h2>
                  <ul className="text-gray-500 dark:text-gray-400 font-medium">
                      <li className="mb-4">
                          <a href="https://flowbite.com/" className="hover:underline">FAQs</a>
                      </li>
                      <li>
                          <a href="https://tailwindcss.com/" className="hover:underline">Contact Us</a>
                      </li>
                  </ul>
              </div>
              <div>
                  <h2 className="mb-6 text-sm font-semibold text-gray-900 uppercase dark:text-white">Legal</h2>
                  <ul className="text-gray-500 dark:text-gray-400 font-medium">
                      <li className="mb-4">
                          <a href="https://flowbite.com/" className="hover:underline">Terms of Service</a>
                      </li>
                      <li>
                          <a href="https://tailwindcss.com/" className="hover:underline">Privacy Policy</a>
                      </li>
                  </ul>
              </div>
             
          </div>
      </div>
     
      <div className="sm:flex sm:items-center sm:justify-between mt-8">
          <span className="text-sm text-gray-500 sm:text-center dark:text-gray-400">
          @Copyright 2024 - Vultisig.
          </span>
         
      </div>
    </div>
  </footer>
    );
  }
  