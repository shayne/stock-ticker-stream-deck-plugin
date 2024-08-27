#!/bin/sh

pid=$(ps -fe | grep '/Applications/Elgato Stream Deck.app/Contents/MacOS/Stream Deck' | awk '{print $2}')
if [[ -n $pid ]]; then
    kill $pid
else
    echo "Does not exist"
fi
