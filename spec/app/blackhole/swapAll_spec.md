# blackhole dex handler



## Overview

Under app/ there are codes which run fiber app server. In there, I want to add new handler which deals with blockchain information, the blackhole dex. 



## Functions



### swap

- description

  - This launches swap all of the token to the other.
- interface definition

  - Http method : POST
  - Reqeust Paramet
    - swapAll : int type

  - Response
    - success or fail message : string
- background

  - There is no implemented structure which launches the swapAll. So just define the interface and use it for the handler.
  - [IMPORTANT] Only implement the handler. Do not implement the swapAll function.



