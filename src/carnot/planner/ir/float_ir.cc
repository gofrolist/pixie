/*
 * Copyright 2018- The Pixie Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "src/carnot/planner/ir/float_ir.h"

namespace px {
namespace carnot {
namespace planner {

Status FloatIR::ToProtoImpl(planpb::ScalarValue* value) const {
  value->set_float64_value(val_);
  return Status::OK();
}

uint64_t FloatIR::HashValueImpl() const {
  return ::util::Hash64(reinterpret_cast<const char*>(&val_), sizeof(double));
}

/* Float IR */
Status FloatIR::Init(double val) {
  val_ = val;
  return Status::OK();
}

Status FloatIR::CopyFromNodeImpl(const IRNode* node, absl::flat_hash_map<const IRNode*, IRNode*>*) {
  const FloatIR* float_ir = static_cast<const FloatIR*>(node);
  val_ = float_ir->val_;
  return Status::OK();
}

}  // namespace planner
}  // namespace carnot
}  // namespace px
