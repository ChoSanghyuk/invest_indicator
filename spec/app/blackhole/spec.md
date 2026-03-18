# blackhole dex handler



## Overview

Under app/ there are codes which run fiber app server. In there, I want to add new handler which deals with blockchain information, the blackhole dex. 



## Functions



### profit

- description

  - This returns the how much profit I've earned from the baseDate.

- interface definition

  - Http method : GET
  - Reqeust Paramet
    - baseDate : Date type

  - Response
    - baseTotalAsset : big int
      - currentTotalAsset : big int
      - profitRate : float
      - profitAmtAvax
      - profitAmtUsdc

- background

  - Refer the AssetSnapshotRecord in internal/db/blackhole_sql.go
  - AssetSnapshotRecord is being recored in every 3 hours or whenver the dex position is changed.
  - Get AssetSnapshotRecord of the baseDate and the most recent one, and compare them to make response.

- furthermore

  - If there is a much better way to manage AssetSnapshotRecord structures in internal/db/blackhole_sql.go, refactor or reposition it in the `internal/model/types.go`.



