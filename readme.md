A simple converter from LittleNavMap's lnmpln format to AWC format for P3D CIVA-A by Marco Ravanello & Gianfranco Corrias.

Usage:
lncivaconv [-1] flightplan.lnmpln  

How it works:
Converter read lnmpln file and create several awc files.
First awc file start from waypoint 2, because waypoint 1 as usual filled by hand on flight preparation phase.
And because it first waypoint in LNM' plan will be skipped by converter.
But if set -1 flag, then first waypoint not will be skipped. It's handy when you create separate files for SID or STAR.

If flight plan contain more than 8 waypoints, converter create several of files with names DEP_DST_x, where is the x number
that the files should load.

Waypoints arranged by files with next rule:
First file contain waypoints from 2 to 9. And when waypoint 9 will be active, you must load second awc file.
Second file contain waypoints from 1 to 8, because active waypoint can't be rewrited.
A third file will be contained waypoints from 9 to 7 through 1. In all next files first loaded waypoint number will be 
smaller by one than in the previous file. 
Example:
1) 1->9 (switch files before 9 waypoint)
2) 1->8 (switch files before 8 waypoint)
3) 9->7 (switch files before 7 waypoint)
4) 8->6 (switch files before 6 waypoint)
5) 7->5 (switch files before 5 waypoint)
and so on...

Also, if flight plan will be contained any VORs, they will be saved in ADC file, for DME corrections.
But without their freqs, because LNM don't save it.

At the end, a TXT file will be created that contains a list of all points with their location in the files. 