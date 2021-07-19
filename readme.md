Simple converter from LittleNavMap's lnmpln format to AWC format for P3D CIVA-A by Marco Ravanello & Gianfranco Corrias.

How it work:
Converter read lnmpln file and create several awc files.
First awc file start from waypoint 2, because waypoint 1 as usual filled by hand on flight preparation phase.
And because it first waypoint in LNM' plan will be skipped by converter.

If flight plan contain more then 8 waypoints, converter create several of files with names DEP_DST_x, where is the x number
that the files should load.

Waypoints arranged by files with next rule:
First file contain waypoints from 2 to 9. And when waypoint 9 will be active, you must load second awc file.
Second file contain waypoints from 1 to 8, because active waypoint can't be rewrited. And when waypoint 8 will be active
you must load next file.
Third and other files will be contain waypoints from 9 to 8 through 1.
