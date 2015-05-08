cd diff
gopherjs build -o out/test.js
cd ..

cd worklog/main_test
gopherjs build -o out/test.js
cd ../../

karma run
