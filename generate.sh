#!/bin/bash
# Copyright 2024 Nokia
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

clang-format -i -style=file:clang-format.style schema.proto
clang-format -i -style=file:clang-format.style data.proto
clang-format -i -style=file:clang-format.style tree_persist.proto

SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

SCHEMA_OUT_DIR=$SCRIPTPATH/sdcpb

mkdir -p $SCHEMA_OUT_DIR
protoc --go_out=paths=source_relative:$SCHEMA_OUT_DIR --go-grpc_out=paths=source_relative:$SCHEMA_OUT_DIR -I $SCRIPTPATH $SCRIPTPATH/schema.proto $SCRIPTPATH/data.proto

mkdir -p $SCRIPTPATH/tree_persist
protoc --go_out=paths=source_relative:./tree_persist -I $SCRIPTPATH $SCRIPTPATH/tree_persist.proto